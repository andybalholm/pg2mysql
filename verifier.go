package pg2mysql

import "fmt"

type Verifier interface {
	Verify() error
}

type verifier struct {
	src, dst DB
	watcher  VerifierWatcher
}

func NewVerifier(src, dst DB, watcher VerifierWatcher) Verifier {
	return &verifier{
		src:     src,
		dst:     dst,
		watcher: watcher,
	}
}

func (v *verifier) Verify() error {
	srcSchema, err := BuildSchema(v.src)
	if err != nil {
		return fmt.Errorf("failed to build source schema: %s", err)
	}

	for _, table := range srcSchema.Tables {
		v.watcher.TableVerificationDidStart(table.Name)

		var missingRows int64
		err = EachMissingRow(v.src, v.dst, table, func(scanArgs []interface{}) {
			missingRows++
		})
		if err != nil {
			v.watcher.TableVerificationDidFinishWithError(table.Name, err)
			continue
		}

		v.watcher.TableVerificationDidFinish(table.Name, missingRows)
	}

	return nil
}
