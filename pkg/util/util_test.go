package util_test

import (
	sqlv1alpha1 "github.com/HamzaZo/terrak8s-operator/api/v1alpha1"
	"github.com/HamzaZo/terrak8s-operator/controllers"
	"github.com/HamzaZo/terrak8s-operator/pkg/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"time"
)

var _ = Describe("Util", func() {
	cr := sqlv1alpha1.PostgreSql{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-instance",
			Namespace: "default",
		},
	}
	var (
		variableMap map[string]interface{}
		json        []byte
		prettyJson  []byte
		fullPath    string
	)
	Context("AddFinalizer", func() {
		It("adds finalizer to Postgresql Resource", func() {
			postgresql := &cr
			util.AddFinalizer(postgresql, controllers.Finalizer)
			Expect(postgresql.GetFinalizers()).To(ContainElement(controllers.Finalizer))
		})
	})

	Context("HasFinalizer", func() {
		It("return true if Postgresql has finalizer", func() {
			postgresql := &cr
			Expect(util.HasFinalizer(postgresql, controllers.Finalizer)).To(BeTrue())

		})

		It("return false if Postgresql does not have finalizer", func() {
			postgresql := &cr
			Expect(util.HasFinalizer(postgresql, "foo-finalizer")).To(BeFalse())

		})
	})

	Context("RemoveFinalizer", func() {
		It("remove finalizer from Postgresql ", func() {
			postgresql := &cr
			util.RemoveFinalizer(postgresql, controllers.Finalizer)
			Expect(postgresql.GetFinalizers()).NotTo(ContainElement(controllers.Finalizer))
		})
	})
	Context("Resource being deleted", func() {
		It("return false if Postgresql deleteTimestamp is equal to zero ", func() {
			postgresql := &cr
			Expect(util.IsBeingDeleted(postgresql)).To(BeFalse())
		})
		It("return true if Postgresql deleteTimestamp isn't equal to zero ", func() {
			postgresql := sqlv1alpha1.PostgreSql{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-instance",
					Namespace: "default",
					DeletionTimestamp: &metav1.Time{
						Time: time.Date(2021, time.May, 5, 5, 5, 5, 0, time.UTC).Local(),
					},
				},
			}
			Expect(util.IsBeingDeleted(&postgresql)).To(BeTrue())
		})
	})
	Context("Validate Password", func() {
		It("return false if password format is not respected ", func() {
			password := "Jmypassword"
			util.IsValidPasswordFormat(password)
			Expect(util.IsValidPasswordFormat(password)).To(BeFalse())
		})
		It("return true if password format is respected ", func() {
			password := "jEnv2000!"
			util.IsValidPasswordFormat(password)
			Expect(util.IsValidPasswordFormat(password)).To(BeTrue())
		})
	})
	Context("create directory", func() {
		It("create directory for tf files ", func() {
			str, err := util.CreateDirectory(cr.Namespace, cr.Name)
			Expect(err).ToNot(HaveOccurred(), "failed to create directory")
			parentDir := os.TempDir()
			Expect(str).To(Equal(parentDir + "/" + cr.Namespace + "_" + cr.Name))

		})
	})
	Context("convert and write tf files", func() {
		BeforeEach(func() {
			variableMap = map[string]interface{}{
				"project": "my-project",
				"region":  "europe-west1",
				"zone":    "europe-west1-b",
			}
			json = []byte(`{"project": "my-project", "region":"europe-west1","zone":"europe-west1-b"}`)
			parentDir := os.TempDir()
			fullPath = parentDir + "/" + cr.Namespace + "_" + cr.Name
		})
		It("return a well formatted json based on tf tag ", func() {
			out, err := util.ToJson(variableMap)
			Expect(err).ToNot(HaveOccurred(), "failed to convert to json")
			prettyJson, err = util.GetPrettyJSON(out)
			Expect(err).ToNot(HaveOccurred(), "failed to convert to pretty json")
			Expect(string(prettyJson)).Should(MatchJSON(json))
		})
		It("should write jsonOutput to file", func() {
			err := util.WriteToFile(prettyJson, fullPath, "file.tf.json")
			Expect(err).ToNot(HaveOccurred(), "failed to write to file")
			Expect(fullPath + "/" + "file.tf.json").Should(BeARegularFile())
		})
		It("clean created directory ", func() {
			err := util.HouseCleaning(fullPath)
			Expect(err).ToNot(HaveOccurred(), "failed to clean")
			Expect(fullPath).ShouldNot(BeADirectory())
		})
	})

})
