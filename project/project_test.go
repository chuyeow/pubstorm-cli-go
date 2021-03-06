package project_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/project"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "project")
}

var _ = Describe("Project", func() {
	DescribeTable("ValidateName()",
		func(name string, expectedError error) {
			proj := &project.Project{Name: name}
			err := proj.ValidateName()

			if expectedError == nil {
				Expect(err).To(BeNil())
			} else {
				Expect(err).To(Equal(expectedError))
			}
		},

		Entry("normal", "abc", nil),
		Entry("allows hyphens", "good-one", nil),
		Entry("allows multiple hyphens", "hello-world--foobar", nil),
		Entry("disallows uppercase characters", "Bad-One", project.ErrNameInvalid),
		Entry("disallows starting with a hyphen", "-abc", project.ErrNameInvalid),
		Entry("disallows ending with a hyphen", "abc-", project.ErrNameInvalid),
		Entry("disallows spaces", "good one", project.ErrNameInvalid),
		Entry("disallows names shorter than 3 characters", "aa", project.ErrNameInvalidLength),
		Entry("disallows names longer than 63 characters", strings.Repeat("a", 64), project.ErrNameInvalidLength),
		Entry("disallows special characters", "good&one", project.ErrNameInvalid),
	)

	DescribeTable("DefaultDomain()",
		func(name, expectedDomain string) {
			origDefaultDomain := config.DefaultDomain
			config.DefaultDomain = "test.dev"
			defer func() {
				config.DefaultDomain = origDefaultDomain
			}()
			proj := &project.Project{Name: name}
			result := proj.DefaultDomain()

			Expect(result).To(Equal(expectedDomain))
		},

		Entry("equals", "aaa", "aaa.test.dev"),
		Entry("equals", "foo-1", "foo-1.test.dev"),
		Entry("equals", "完成", "完成.test.dev"),
	)

	Describe("file system dependent tests", func() {
		var (
			currDir string
			tempDir string
			err     error
		)

		BeforeEach(func() {
			currDir, err = os.Getwd()
			Expect(err).To(BeNil())
			tempDir, err = ioutil.TempDir("", "rise-test")
			Expect(err).To(BeNil())
			os.Chdir(tempDir)
		})

		AfterEach(func() {
			os.Chdir(currDir)
			os.RemoveAll(tempDir)
		})

		Describe("ValidatePath()", func() {
			var tempDeployDir string

			BeforeEach(func() {
				tempDeployDir = filepath.Join(tempDir, "public")
				err = os.Mkdir(tempDeployDir, 0700)
				Expect(err).To(BeNil())
			})

			Context("when path is absolute", func() {
				It("returns error", func() {
					proj := &project.Project{Path: tempDeployDir}
					Expect(proj.ValidatePath()).To(Equal(project.ErrPathNotRelative))
				})
			})

			Context("when path does not exist", func() {
				It("returns error", func() {
					proj := &project.Project{Path: "./public2"}
					Expect(proj.ValidatePath()).To(Equal(project.ErrPathNotExist))
				})
			})

			Context("when path is not a directory", func() {
				It("returns error", func() {
					err = ioutil.WriteFile(filepath.Join(tempDir, "public2"), []byte{'a'}, 0600)
					Expect(err).To(BeNil())

					proj := &project.Project{Path: "./public2"}
					Expect(proj.ValidatePath()).To(Equal(project.ErrPathNotDir))
				})
			})

			Context("when path is relative and exists", func() {
				It("returns nil", func() {
					proj := &project.Project{Path: "./public"}
					Expect(proj.ValidatePath()).To(BeNil())
				})
			})
		})

		Describe("Save()", func() {
			It("persists only name and path in project json file in the current working directory", func() {
				proj := &project.Project{
					Name:                 "foo-bar-express",
					Path:                 "./build",
					DefaultDomainEnabled: true,
					ForceHTTPS:           false,
				}

				err = proj.Save()
				Expect(err).To(BeNil())

				f, err := os.Open(filepath.Join(tempDir, config.ProjectJSON))
				Expect(err).To(BeNil())
				defer f.Close()

				var j map[string]interface{}
				err = json.NewDecoder(f).Decode(&j)
				Expect(err).To(BeNil())

				Expect(j).NotTo(BeNil())
				Expect(j).To(Equal(map[string]interface{}{
					"name": "foo-bar-express",
					"path": "./build",
				}))
			})
		})

		Describe("Load()", func() {
			Context("when the project json does not exist", func() {
				It("returns error", func() {
					proj, err := project.Load()
					Expect(err).NotTo(BeNil())
					Expect(os.IsNotExist(err)).To(BeTrue())
					Expect(proj).To(BeNil())
				})
			})

			Context("when the project json exists", func() {
				BeforeEach(func() {
					err = ioutil.WriteFile(config.ProjectJSON, []byte(`
						{
							"name": "good-beer-company",
							"path": "./output",
							"enable_stats": true,
							"force_https": true
						}
					`), 0600)
					Expect(err).To(BeNil())
				})

				It("loads the project json and returns a project", func() {
					proj, err := project.Load()
					Expect(err).To(BeNil())

					Expect(proj).NotTo(BeNil())
					Expect(proj.Name).To(Equal("good-beer-company"))
					Expect(proj.Path).To(Equal("./output"))
				})

				It("ignores fields other than the project name and path", func() {
					proj, err := project.Load()
					Expect(err).To(BeNil())

					Expect(proj).NotTo(BeNil())
					Expect(proj.ForceHTTPS).To(BeFalse())
				})
			})
		})

		Describe("LoadDefault()", func() {
			Context("when pubstorm.default.json does not exist", func() {
				It("returns a zero value pointer to a project", func() {
					proj, err := project.LoadDefault()
					Expect(err).To(BeNil())
					Expect(proj).To(Equal(&project.Project{}))
				})
			})

			Context("when a pubstorm.default.json file exists", func() {
				BeforeEach(func() {
					err = ioutil.WriteFile(config.DefaultProjectJSONPath, []byte(`
						{
							"path": "./_site"
						}
					`), 0600)
					Expect(err).To(BeNil())
				})

				It("loads the project from it", func() {
					proj, err := project.LoadDefault()
					Expect(err).To(BeNil())

					Expect(proj).NotTo(BeNil())
					Expect(proj.Path).To(Equal("./_site"))
				})
			})
		})

		Describe("Delete()", func() {
			var proj *project.Project

			BeforeEach(func() {
				err = ioutil.WriteFile(config.ProjectJSON, []byte(`
					{
						"name": "good-beer-company",
						"path": "./output",
						"enable_stats": false,
						"force_https": true
					}
				`), 0600)
				Expect(err).To(BeNil())

				var err error
				proj, err = project.Load()
				Expect(err).To(BeNil())
			})

			It("deletes the project json file", func() {
				err := proj.Delete()
				Expect(err).To(BeNil())

				_, err = os.Stat(config.ProjectJSON)
				Expect(os.IsNotExist(err)).To(BeTrue())
			})
		})
	})
})
