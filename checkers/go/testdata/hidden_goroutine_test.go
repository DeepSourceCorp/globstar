package golang

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ana "globstar.dev/analysis"
)

func parseGoCode(t *testing.T, source []byte) *ana.ParseResult {
	pass, err := ana.Parse("", source, ana.LangGo, ana.LangGo.Grammar())
	require.NoError(t, err)

	return pass
}

func TestHiddenGoroutine(t *testing.T) {
	t.Run("get program structure", func(t *testing.T) {
		source := []byte(`
		func HiddenGoRoutine(){
			// Hello
			// fmt.Println("Hola")
			go func(){
				fmt.Println("Hidden goroutine")
			}()
		}
		`)
		parseResult := parseGoCode(t, source)
		pass := &ana.Pass{
			Analyzer:    HiddenGoRoutineAnalyzer,
			FileContext: parseResult,
		}
		_, err := detectHiddenGoroutine(pass)
		assert.Nil(t, err)
	})
}
