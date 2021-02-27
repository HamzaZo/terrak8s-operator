package util

import (
	"bytes"
	"encoding/json"
	sqlv1alpha1 "github.com/HamzaZo/terrak8s-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	kubeApiMetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

var (
	log           logr.Logger
	isValidLength = false
	isUpperChar   = false
	isLowerChar   = false
	isNumber      = false
	isSpecialChar = false
)

const (
	passwordMinLength = 7
)

//GetPrettyJSON return a pretty json format
func GetPrettyJSON(byteJson []byte) ([]byte, error) {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, byteJson, "", "  ")
	if err != nil {
		return nil, err
	}
	return prettyJSON.Bytes(), err
}

//ToJson return json based on tf tag
func ToJson(content interface{}) ([]byte, error) {
	jsonit := jsoniter.Config{
		EscapeHTML:             true,
		SortMapKeys:            true,
		ValidateJsonRawMessage: true,
		TagKey:                 "tf",
	}.Froze()

	val, err := jsonit.MarshalIndent(content, "", " ")
	if err != nil {
		return nil, err
	}
	return val, nil
}

//CreateDirectory create a directory to host generated tf based on namespace and instance name
func CreateDirectory(namespace string, instance string) (string, error) {
	var fullName string
	paths := []string{"instance", "bucket"}
	parentDir := os.TempDir()
	pathName := parentDir + "/" + namespace + "_" + instance
	for _, k := range paths {
		fullName = filepath.Join(pathName, k)
		_, err := os.Stat(fullName)
		if os.IsNotExist(err) {
			err = os.MkdirAll(fullName, os.ModePerm)
			if err != nil {
				log.Error(err, "unable to create directory")
				return "", err
			}
		}
	}

	return pathName, nil
}

func WriteToFile(b []byte, path string, name string) error {
	if err := ioutil.WriteFile(path+"/"+name, b, 0755); err != nil {
		return err
	}
	return nil
}

//HouseCleaning clean what have been created
func HouseCleaning(dirPath string) error {
	err := os.RemoveAll(dirPath)
	if err != nil {
		return err
	}
	errs := os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
	if errs != nil {
		return errs
	}
	return nil
}

// HasFinalizer returns whether this object has the passed finalizer
func HasFinalizer(obj kubeApiMetav1.Object, finalizer string) bool {
	for _, fin := range obj.GetFinalizers() {
		if fin == finalizer {
			return true
		}
	}
	return false
}

// IsBeingDeleted returns whether this object has been requested to be deleted
func IsBeingDeleted(obj kubeApiMetav1.Object) bool {
	return !obj.GetDeletionTimestamp().IsZero()
}

// AddFinalizer adds the passed finalizer this object
func AddFinalizer(obj kubeApiMetav1.Object, finalizer string) {
	if !HasFinalizer(obj, finalizer) {
		obj.SetFinalizers(append(obj.GetFinalizers(), finalizer))
	}
}

// RemoveFinalizer removes the passed finalizer from object
func RemoveFinalizer(obj kubeApiMetav1.Object, finalizer string) {
	for i, fin := range obj.GetFinalizers() {
		if fin == finalizer {
			finalizers := obj.GetFinalizers()
			finalizers[i] = finalizers[len(finalizers)-1]
			obj.SetFinalizers(finalizers[:len(finalizers)-1])
			return
		}
	}
}

//UpdateOutput Unmarshal the returned output and update the output status
func UpdateOutput(instanceOutput *sqlv1alpha1.PostgreSql, output string) (error, *sqlv1alpha1.PostgreSql) {
	out := extractOutput(output)
	jsonit := jsoniter.Config{
		EscapeHTML:             true,
		SortMapKeys:            true,
		ValidateJsonRawMessage: true,
		TagKey:                 "json",
	}.Froze()

	var d map[string]interface{}
	finalMap := make(map[string]interface{})

	err := jsonit.Unmarshal(out, &d)
	if err != nil {
		return err, nil
	}
	for e, v := range d {
		m := v.(map[string]interface{})
		for i, k := range m {
			if i == "value" {
				finalMap[e] = k
			}
		}
	}

	err = mapstructure.Decode(finalMap, &instanceOutput.Status.Output)
	if err != nil {
		return err, nil
	}
	return nil, instanceOutput
}

//extractOutput return the content of output
func extractOutput(output string) []byte {
	finder := "{"
	lengthOfFinder := len(finder)
	indexOf := strings.Index(output, finder)
	var finalContent string
	if indexOf != -1 {
		// If found get content
		indexOf += lengthOfFinder
		end := "Ini"
		if strings.Contains(output[indexOf:], end) {
			y := strings.Index(output[indexOf:], end)
			//10 is the number of the left bytes, while extracting the content of output
			//it's a dirty workaround to avoid error while preforming Unmarshal of finalContent
			finalContent = output[indexOf : indexOf+y-10]
		} else {
			finalContent = output[indexOf:]
		}
	} else {
		// If not content after empty.
		finalContent = ""
	}
	res := []byte(finalContent)
	out := []byte(`{`)
	out = append(out, res...)

	return out
}

//IsValidPasswordFormat Validate password format
func IsValidPasswordFormat(password string) bool {
	if len(password) >= passwordMinLength {
		isValidLength = true
	}
	for _, s := range password {
		switch {
		case unicode.IsUpper(s):
			isUpperChar = true
		case unicode.IsLower(s):
			isLowerChar = true
		case unicode.IsNumber(s):
			isNumber = true
		case unicode.IsPunct(s) || unicode.IsSymbol(s):
			isSpecialChar = true
		}
	}
	return isValidLength && isUpperChar && isLowerChar && isNumber && isSpecialChar
}
