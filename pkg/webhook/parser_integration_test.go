/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package webhook_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	admissionreg "k8s.io/api/admissionregistration/v1beta1"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/controller-tools/pkg/webhook"
)

var _ = Describe("Webhook Generation From Parsing to CustomResourceDefinition", func() {
	var (
		err       error
		cwd       string
		outputDir string
		reg       *markers.Registry
		pkgs      []*loader.Package
	)

	BeforeEach(func() {
		// TODO(directxman12): test generation across multiple versions (right
		// now, we're trusting k/k's conversion code, though, which is probably
		// fine for the time being)
		By("switching into testdata to appease go modules")
		cwd, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir("./testdata")).To(Succeed()) // go modules are directory-sensitive

		By("loading the roots")
		pkgs, err = loader.LoadRoots(".")
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))

		By("setting up the parser")
		reg = &markers.Registry{}
		Expect(reg.Register(webhook.ConfigDefinition)).To(Succeed())

		By("requesting that the manifest be generated")
		outputDir, err = ioutil.TempDir("", "webhook-integration-test")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if err = os.RemoveAll(outputDir); err != nil {
			fmt.Fprintf(GinkgoWriter, "warning: failed to remove test output directory %s: %v", outputDir, err)
		}
		if err = os.Chdir(cwd); err != nil {
			fmt.Fprintf(GinkgoWriter, "warning: failed to change directory back to %s: %v", cwd, err)
		}
	})

	It("should properly generate the v1 and v1beta1 webhook definitions", func() {
		g := webhook.Generator{WebhookVersions: []string{"v1", "v1beta1"}}
		ctx := &genall.GenerationContext{
			Collector:  &markers.Collector{Registry: reg},
			Roots:      pkgs,
			OutputRule: genall.OutputToDirectory(outputDir),
		}
		ExpectWithOffset(1, g.Generate(ctx)).To(Succeed())
		v1file, v1beta1file := "manifests.yaml", "manifests.v1beta1.yaml"
		checkOutput(path.Join(outputDir, v1file), v1file)
		checkOutput(path.Join(outputDir, v1beta1file), v1beta1file)
	})
})

func checkOutput(actualPath, expectedPath string) {
	By("loading the generated YAML")
	actualFile, err := ioutil.ReadFile(actualPath)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	actualMutating, actualValidating := unmarshalBoth(actualFile)

	By("loading the desired YAML")
	expectedFile, err := ioutil.ReadFile(expectedPath)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	expectedMutating, expectedValidating := unmarshalBoth(expectedFile)

	By("comparing the two")
	assertSame(actualMutating, expectedMutating)
	assertSame(actualValidating, expectedValidating)
}

// assertSame compares two webhooks.
func assertSame(actual, expected interface{}) {
	ExpectWithOffset(1, actual).To(Equal(expected),
		"type not as expected, check pkg/webhook/testdata/README.md for more details."+
			"\n\nDiff:\n\n%s", cmp.Diff(actual, expected))
}

func unmarshalBoth(in []byte) (mutating admissionreg.MutatingWebhookConfiguration, validating admissionreg.ValidatingWebhookConfiguration) {
	documents := bytes.Split(in, []byte("\n---\n"))[1:]
	ExpectWithOffset(1, documents).To(HaveLen(2), "expected two documents in file, found %d", len(documents))

	ExpectWithOffset(1, yaml.UnmarshalStrict(documents[0], &mutating)).To(Succeed(), "expected the first document in the file to be a mutating webhook configuration")
	ExpectWithOffset(1, yaml.UnmarshalStrict(documents[1], &validating)).To(Succeed(), "expected the second document in the file to be a validating webhook configuration")
	return
}
