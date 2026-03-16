// SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pixiinstall_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"

	executablefakes "github.com/paketo-buildpacks/python-package-managers-run/pkg/executable/fakes"
	pixiinstall "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/pixi"
	summerfakes "github.com/paketo-buildpacks/python-package-managers-run/pkg/summer/fakes"
)

func testPixiRunner(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir    string
		pixiLayerPath string
		pixiCachePath string

		executable *executablefakes.Executable
		executions []pexec.Execution
		summer     *summerfakes.Summer
		runner     pixiinstall.PixiRunner
		buffer     *bytes.Buffer
		logger     scribe.Emitter
	)

	it.Before(func() {
		workingDir = t.TempDir()
		layersDir := t.TempDir()

		pixiLayerPath = filepath.Join(layersDir, "a-pixi-layer")
		pixiCachePath = filepath.Join(layersDir, "a-pixi-cache-path")

		executable = &executablefakes.Executable{}
		executions = []pexec.Execution{}
		executable.ExecuteCall.Stub = func(ex pexec.Execution) error {
			executions = append(executions, ex)
			Expect(os.MkdirAll(filepath.Join(pixiLayerPath, "pixi-meta"), os.ModePerm)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(pixiLayerPath, "pixi-meta", "history"), []byte("some content"), os.ModePerm)).To(Succeed())
			_, err := fmt.Fprintln(ex.Stdout, "stdout output")
			Expect(err).NotTo(HaveOccurred())
			_, err = fmt.Fprintln(ex.Stderr, "stderr output")
			Expect(err).NotTo(HaveOccurred())
			return nil
		}

		summer = &summerfakes.Summer{}
		buffer = bytes.NewBuffer(nil)
		logger = scribe.NewEmitter(buffer)
		runner = pixiinstall.NewPixiRunner(executable, summer, logger)
	})

	context("ShouldRun", func() {
		it("returns true, with no sha, and no error when no lockfile is present", func() {
			run, sha, err := runner.ShouldRun(workingDir, map[string]interface{}{})
			Expect(run).To(BeTrue())
			Expect(sha).To(Equal(""))
			Expect(err).NotTo(HaveOccurred())
		})

		context("when there is an error checking if a lockfile is present", func() {
			it.Before(func() {
				Expect(os.Chmod(workingDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
			})

			it("returns false, with no sha, and an error", func() {
				run, sha, err := runner.ShouldRun(workingDir, map[string]interface{}{})
				Expect(run).To(BeFalse())
				Expect(sha).To(Equal(""))
				Expect(err).To(HaveOccurred())
			})
		})

		context("when a lockfile is present", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(workingDir, pixiinstall.LockfileName), nil, os.ModePerm)).To(Succeed())
			})
			context("and the lockfile sha is unchanged", func() {
				it("return false, with the existing sha, and no error", func() {
					summer.SumCall.Returns.String = "a-sha"
					Expect(os.WriteFile(filepath.Join(workingDir, pixiinstall.LockfileName), nil, os.ModePerm)).To(Succeed())

					metadata := map[string]interface{}{
						"lockfile-sha": "a-sha",
					}

					run, sha, err := runner.ShouldRun(workingDir, metadata)
					Expect(run).To(BeFalse())
					Expect(sha).To(Equal("a-sha"))
					Expect(err).NotTo(HaveOccurred())
				})
				context("and there is and error summing the lock file", func() {
					it.Before(func() {
						summer.SumCall.Returns.Error = errors.New("summing lockfile failed")
					})

					it("returns false, with no sha, and an error", func() {
						run, sha, err := runner.ShouldRun(workingDir, map[string]interface{}{})
						Expect(run).To(BeFalse())
						Expect(sha).To(Equal(""))
						Expect(err).To(MatchError("summing lockfile failed"))

					})
				})
			})

			it("returns true, with a new sha, and no error when the lockfile has changed", func() {
				summer.SumCall.Returns.String = "a-new-sha"
				metadata := map[string]interface{}{
					"lockfile-sha": "a-sha",
				}

				run, sha, err := runner.ShouldRun(workingDir, metadata)
				Expect(run).To(BeTrue())
				Expect(sha).To(Equal("a-new-sha"))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
	context("Execute", func() {
		context("when a lockfile exists", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(workingDir, pixiinstall.LockfileName), nil, os.ModePerm)).To(Succeed())
			})

			it("runs pixi create with the cache layer available in the environment", func() {
				err := runner.Execute(pixiLayerPath, pixiCachePath, workingDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(executions[0].Args).To(Equal([]string{
					"exec",
					"pixi-pack",
					"--use-cache", pixiCachePath,
					"--output-file", "/tmp/project.tar.gz",
					workingDir,
				}))
				Expect(executable.ExecuteCall.CallCount).To(Equal(2))
				Expect(executions[1].Args).To(Equal([]string{
					"exec",
					"pixi-unpack",
					"--output-directory", pixiLayerPath,
					"--env-name", pixiinstall.PixiEnvironmentName,
					"/tmp/project.tar.gz",
				}))
			})
		})

		context("failure cases", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(workingDir, pixiinstall.LockfileName), nil, os.ModePerm)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(workingDir, pixiinstall.ProjectFilename), nil, os.ModePerm)).To(Succeed())
			})
			context("when the pixi exec command fails to run", func() {
				it.Before(func() {
					executable.ExecuteCall.Stub = func(ex pexec.Execution) error {
						_, err := fmt.Fprintln(ex.Stdout, "pixi error stdout")
						Expect(err).NotTo(HaveOccurred())
						_, err = fmt.Fprintln(ex.Stderr, "pixi error stderr")
						Expect(err).NotTo(HaveOccurred())
						return errors.New("some pixi failure")
					}
				})

				it("returns an error and logs the stdout and stderr output from the command", func() {
					err := runner.Execute(pixiLayerPath, pixiCachePath, workingDir)
					Expect(err).To(MatchError("failed to run pixi command: some pixi failure"))
					Expect(buffer.String()).To(ContainLines(
						fmt.Sprintf(
							"    Running 'pixi exec pixi-pack --use-cache %s --output-file /tmp/project.tar.gz %s'",
							pixiCachePath,
							workingDir,
						),
						"      pixi error stdout",
						"      pixi error stderr",
					))
				})
			})

		})
	})
}
