package reader

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ReadCSV(t *testing.T) {
	testCases := []struct {
		name         string
		filePath     string
		lineNum      int
		fileOpenFunc func(name string) (*os.File, error)
		wantContent  interface{}
		wantErr      error
	}{
		/*{
			name:        "os.Stat returns an error",
			filePath:    "random.csv",
			lineNum:     10,
			wantContent: nil,
			wantErr:     errors.New("no such file or directory"),
		},*/
		{
			name:        "os.Stat returns isDir as true",
			filePath:    "../sample",
			lineNum:     10,
			wantContent: nil,
			wantErr:     errors.New("the provided path is a directory"),
		},
		{
			name:     "file open error",
			filePath: "../sample/customers-100.csv",
			lineNum:  10,
			fileOpenFunc: func(name string) (*os.File, error) {
				return nil, errors.New("fake error")
			},
			wantContent: nil,
			wantErr:     errors.New("the provided path is a directory"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			backupFileOpenFunc := osFileOpenFunc

			if testCase.fileOpenFunc != nil {
				osFileOpenFunc = testCase.fileOpenFunc
			}
			defer func() {
				osFileOpenFunc = backupFileOpenFunc
			}()
			
			actualAny, actualError := ReadCSV(testCase.filePath, testCase.lineNum)
			assert.Equal(t, testCase.wantContent, actualAny)
			assert.Equal(t, testCase.wantErr, actualError)
			/*if testCase.wantErr != nil {
				assert.NotNil(t, actualError)
			}*/
		})
	}
}
