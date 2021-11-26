package vault

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/hashicorp/vault/api"
	. "github.com/onsi/gomega"

	v1beta1 "github.com/DoodleScheduling/k8svault-controller/api/v1beta1"
)

type testMapper struct {
	forceApply bool
	path       string
	fields     []v1beta1.FieldMapping
}

func (m *testMapper) IsForceApply() bool {
	return m.forceApply
}

func (m *testMapper) GetPath() string {
	return m.path
}

func (m *testMapper) GetFieldMapping() []v1beta1.FieldMapping {
	return m.fields
}

type testResult struct {
	err    error
	secret *api.Secret
}

type mockReadWriter struct {
	readResult  testResult
	writeResult testResult
	writtenPath string
	writtenData map[string]interface{}
}

func (rw *mockReadWriter) Read(path string) (*api.Secret, error) {
	return rw.readResult.secret, rw.readResult.err
}

func (rw *mockReadWriter) Write(path string, data map[string]interface{}) (*api.Secret, error) {
	fmt.Printf("WRITE %#v -- %#v -- %#v\n", path, data, rw.writeResult.err)
	rw.writtenPath = path
	rw.writtenData = data

	return rw.writeResult.secret, rw.writeResult.err
}

func TestWrite(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name          string
		writeData     map[string]interface{}
		mapper        *testMapper
		readWriter    *mockReadWriter
		expectWritten bool
		expectError   error
		expectData    map[string]interface{}
	}{
		{
			name: "write to /food if path does not exists",
			mapper: &testMapper{
				forceApply: false,
				path:       "/food",
				fields:     []v1beta1.FieldMapping{},
			},
			writeData: map[string]interface{}{
				"fruit": "banana",
			},
			readWriter: &mockReadWriter{
				readResult: testResult{
					err: nil,
					secret: &api.Secret{
						Data: nil,
					},
				},
				writeResult: testResult{
					err:    nil,
					secret: &api.Secret{},
				},
			},
			expectWritten: true,
			expectError:   nil,
			expectData: map[string]interface{}{
				"fruit": "banana",
			},
		},
		{
			name: "Don't overwrite field if path already exists and has the same field",
			mapper: &testMapper{
				forceApply: false,
				path:       "/food",
				fields:     []v1beta1.FieldMapping{},
			},
			writeData: map[string]interface{}{
				"fruit": "banana",
			},
			readWriter: &mockReadWriter{
				readResult: testResult{
					err: nil,
					secret: &api.Secret{
						Data: map[string]interface{}{
							"fruit": "strawberry",
						},
					},
				},
				writeResult: testResult{
					err:    nil,
					secret: &api.Secret{},
				},
			},
			expectWritten: false,
			expectError:   nil,
			expectData:    nil,
		},
		{
			name: "Add field to existing path",
			mapper: &testMapper{
				forceApply: false,
				path:       "/food",
				fields:     []v1beta1.FieldMapping{},
			},
			writeData: map[string]interface{}{
				"fruit": "banana",
			},
			readWriter: &mockReadWriter{
				readResult: testResult{
					err: nil,
					secret: &api.Secret{
						Data: map[string]interface{}{
							"vegtable": "tomato?",
						},
					},
				},
				writeResult: testResult{
					err:    nil,
					secret: &api.Secret{},
				},
			},
			expectWritten: true,
			expectError:   nil,
			expectData: map[string]interface{}{
				"fruit":    "banana",
				"vegtable": "tomato?",
			},
		},
		{
			name: "Add only selected fields from given data to existing path",
			mapper: &testMapper{
				forceApply: false,
				path:       "/food",
				fields: []v1beta1.FieldMapping{
					{Name: "vegetable"},
					{Name: "sweet"},
				},
			},
			writeData: map[string]interface{}{
				"fruit":     "banana",
				"vegetable": "tomato?",
				"sweet":     "icecream",
			},
			readWriter: &mockReadWriter{
				readResult: testResult{
					err: nil,
					secret: &api.Secret{
						Data: map[string]interface{}{
							"carbs": "pasta",
						},
					},
				},
				writeResult: testResult{
					err:    nil,
					secret: &api.Secret{},
				},
			},
			expectWritten: true,
			expectError:   nil,
			expectData: map[string]interface{}{
				"carbs":     "pasta",
				"vegetable": "tomato?",
				"sweet":     "icecream",
			},
		},
		{
			name: "Add only selected fields from given data to existing path which has partially the same fields",
			mapper: &testMapper{
				forceApply: false,
				path:       "/food",
				fields: []v1beta1.FieldMapping{
					{Name: "vegetable"},
					{Name: "sweet"},
				},
			},
			writeData: map[string]interface{}{
				"fruit":     "banana",
				"vegetable": "cucumber",
				"sweet":     "icecream",
			},
			readWriter: &mockReadWriter{
				readResult: testResult{
					err: nil,
					secret: &api.Secret{
						Data: map[string]interface{}{
							"vegetable": "tomato?",
							"carbs":     "pasta",
						},
					},
				},
				writeResult: testResult{
					err:    nil,
					secret: &api.Secret{},
				},
			},
			expectWritten: true,
			expectError:   nil,
			expectData: map[string]interface{}{
				"carbs":     "pasta",
				"vegetable": "tomato?",
				"sweet":     "icecream",
			},
		},
		{
			name: "Add only selected fields from given data to existing path and rename one",
			mapper: &testMapper{
				forceApply: false,
				path:       "/food",
				fields: []v1beta1.FieldMapping{
					{Name: "vegetable", Rename: "fruit"},
					{Name: "carbs"},
				},
			},
			writeData: map[string]interface{}{
				"vegetable": "tomato?",
				"carbs":     "pasta",
				"dessert":   "icecream",
			},
			readWriter: &mockReadWriter{
				readResult: testResult{
					err: nil,
					secret: &api.Secret{
						Data: map[string]interface{}{
							"vegetable": "cucumber",
						},
					},
				},
				writeResult: testResult{
					err:    nil,
					secret: &api.Secret{},
				},
			},
			expectWritten: true,
			expectError:   nil,
			expectData: map[string]interface{}{
				"vegetable": "cucumber",
				"fruit":     "tomato?",
				"carbs":     "pasta",
			},
		},
		{
			name: "Add only selected fields from given data and overwrite existing fields",
			mapper: &testMapper{
				forceApply: true,
				path:       "/food",
				fields: []v1beta1.FieldMapping{
					{Name: "vegetable", Rename: "fruit"},
					{Name: "carbs"},
				},
			},
			writeData: map[string]interface{}{
				"vegetable": "tomato?",
				"carbs":     "pasta",
				"dessert":   "icecream",
			},
			readWriter: &mockReadWriter{
				readResult: testResult{
					err: nil,
					secret: &api.Secret{
						Data: map[string]interface{}{
							"vegetable": "cucumber",
							"carbs":     "rise",
						},
					},
				},
				writeResult: testResult{
					err:    nil,
					secret: &api.Secret{},
				},
			},
			expectWritten: true,
			expectError:   nil,
			expectData: map[string]interface{}{
				"vegetable": "cucumber",
				"fruit":     "tomato?",
				"carbs":     "pasta",
			},
		},
		{
			name: "Fails if mapping declares a field which is not found",
			mapper: &testMapper{
				forceApply: false,
				path:       "/food",
				fields: []v1beta1.FieldMapping{
					{Name: "fruit"},
				},
			},
			writeData: map[string]interface{}{
				"vegetable": "tomato?",
			},
			readWriter: &mockReadWriter{
				readResult: testResult{
					err: nil,
					secret: &api.Secret{
						Data: nil,
					},
				},
				writeResult: testResult{
					err:    nil,
					secret: &api.Secret{},
				},
			},
			expectWritten: false,
			expectError:   ErrFieldNotAvailable,
			expectData: map[string]interface{}{
				"vegetable": "cucumber",
				"fruit":     "tomato?",
				"carbs":     "pasta",
			},
		},
		{
			name: "return error if read fails",
			mapper: &testMapper{
				forceApply: false,
				path:       "/food",
				fields:     []v1beta1.FieldMapping{},
			},
			writeData: map[string]interface{}{
				"fruit": "banana",
			},
			readWriter: &mockReadWriter{
				readResult: testResult{
					err:    errors.New("read fails"),
					secret: nil,
				},
				writeResult: testResult{
					err:    nil,
					secret: &api.Secret{},
				},
			},
			expectWritten: false,
			expectError:   errors.New("read fails"),
			expectData:    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := &VaultHandler{
				logger: logr.Discard(),
				c:      test.readWriter,
			}

			written, err := handler.Write(test.mapper, test.writeData)
			if test.expectError == nil {
				g.Expect(err).NotTo(HaveOccurred(), "write error occurd but should not")
			} else {
				g.Expect(err).To(Equal(test.expectError))
			}

			g.Expect(written).To(Equal(test.expectWritten))

			if test.expectWritten == true {
				g.Expect(test.readWriter.writtenData).To(Equal(test.expectData))
				g.Expect(test.readWriter.writtenPath).To(Equal(test.mapper.GetPath()))
			}
		})
	}
}
