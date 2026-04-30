// SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package uvinstall_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/sclevine/spec"

	executablefakes "github.com/paketo-buildpacks/python-package-managers-run/pkg/executable/fakes"
	uv "github.com/paketo-buildpacks/python-package-managers-run/pkg/packagers/uv"
	summerfakes "github.com/paketo-buildpacks/python-package-managers-run/pkg/summer/fakes"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testUvRunner(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir  string
		uvLayerPath string
		uvCachePath string

		executable *executablefakes.Executable
		executions []pexec.Execution
		summer     *summerfakes.Summer
		runner     uv.UvRunner
		buffer     *bytes.Buffer
		logger     scribe.Emitter
	)

	it.Before(func() {
		workingDir = t.TempDir()
		layersDir := t.TempDir()

		uvLayerPath = filepath.Join(layersDir, "a-uv-layer")
		uvCachePath = filepath.Join(layersDir, "a-uv-cache-path")

		executable = &executablefakes.Executable{}
		executions = []pexec.Execution{}
		executable.ExecuteCall.Stub = func(ex pexec.Execution) error {
			executions = append(executions, ex)
			Expect(os.MkdirAll(filepath.Join(uvLayerPath, "uv-meta"), os.ModePerm)).To(Succeed())
			// For reasons currently unknown, the search call triggers a permission issue in the tests
			Expect(os.Chmod(filepath.Join(uvLayerPath, "uv-meta"), os.ModePerm)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(uvLayerPath, "uv-meta", "history"), []byte("some content"), os.ModePerm)).To(Succeed())
			_, err := fmt.Fprintln(ex.Stdout, "stdout output")
			Expect(err).NotTo(HaveOccurred())
			_, err = fmt.Fprintln(ex.Stderr, "stderr output")
			Expect(err).NotTo(HaveOccurred())
			return nil
		}

		summer = &summerfakes.Summer{}
		buffer = bytes.NewBuffer(nil)
		logger = scribe.NewEmitter(buffer)
		runner = uv.NewUvRunner(executable, summer, logger)
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
				Expect(os.WriteFile(filepath.Join(workingDir, uv.LockfileName), nil, os.ModePerm)).To(Succeed())
			})
			context("and the lockfile sha is unchanged", func() {
				it("return false, with the existing sha, and no error", func() {
					summer.SumCall.Returns.String = "a-sha"
					Expect(os.WriteFile(filepath.Join(workingDir, uv.LockfileName), nil, os.ModePerm)).To(Succeed())

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
					uv.LockfileShaName: "a-sha",
				}

				run, sha, err := runner.ShouldRun(workingDir, metadata)
				Expect(run).To(BeTrue())
				Expect(sha).To(Equal("a-new-sha"))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	context("Execute", func() {
		context("when a vendor dir is present", func() {
			var vendorPath string
			it.Before(func() {
				vendorPath = filepath.Join(workingDir, "vendor")
				Expect(os.Mkdir(vendorPath, os.ModePerm))
				Expect(os.WriteFile(filepath.Join(workingDir, uv.LockfileName), nil, os.ModePerm)).To(Succeed())
			})

			it("runs uv sync with additional vendor venv and WITHOUT cache layer args", func() {
				err := runner.Execute(uvLayerPath, uvCachePath, workingDir)
				Expect(err).NotTo(HaveOccurred())

				args := []string{
					"sync",
					"--no-index",
				}

				Expect(executions[0].Args).To(Equal(args))
				venvPath := filepath.Join(uvLayerPath, "venv")
				Expect(executions[0].Env).To(ContainElement(fmt.Sprintf("HOME=%s", uvLayerPath)))
				Expect(executions[0].Env).To(ContainElement(fmt.Sprintf("VIRTUAL_ENV=%s", venvPath)))
				Expect(executions[0].Env).To(ContainElement(fmt.Sprintf("UV_PROJECT_ENVIRONMENT=%s", venvPath)))
				Expect(executions[0].Env).To(ContainElement(fmt.Sprintf("UV_WORKING_DIR=%s", workingDir)))
				Expect(executions[0].Env).To(ContainElement("UV_PYTHON=/layers/paketo-buildpacks_cpython/cpython/bin/python"))
				Expect(executions[0].Env).To(ContainElement("UV_OFFLINE=1"))
				Expect(executions[0].Env).To(ContainElement("LD_LIBRARY_PATH=/layers/paketo-buildpacks_cpython/cpython/lib"))
				userFindLinks, _ := os.LookupEnv("BP_UV_FIND_LINKS")
				findLinks, _ := os.LookupEnv("UV_FIND_LINKS")
				combinedFindLinks := []string{userFindLinks, findLinks}
				combinedFindLinks = append(combinedFindLinks, vendorPath)
				Expect(executions[0].Env).To(ContainElement(fmt.Sprintf("UV_FIND_LINKS=%s", strings.TrimLeft(strings.Join(combinedFindLinks, " "), " "))))
				Expect(executable.ExecuteCall.CallCount).To(Equal(1))
				Expect(buffer.String()).To(ContainLines(
					fmt.Sprintf("    Running 'uv %s'", strings.Join(args, " ")),
					"      stdout output",
					"      stderr output",
				))
			})

			context("failure cases", func() {
				context("when there is an error running the uv command", func() {
					it.Before(func() {
						executable.ExecuteCall.Stub = func(ex pexec.Execution) error {
							_, err := fmt.Fprintln(ex.Stdout, "uv error stdout")
							Expect(err).NotTo(HaveOccurred())
							_, err = fmt.Fprintln(ex.Stderr, "uv error stderr")
							Expect(err).NotTo(HaveOccurred())
							return errors.New("some uv failure")
						}
					})

					it("returns an error with stdout/stderr output", func() {
						err := runner.Execute(uvLayerPath, uvCachePath, workingDir)
						Expect(err).To(MatchError("failed to run uv command: some uv failure"))

						args := []string{
							"sync",
							"--no-index",
						}
						Expect(buffer.String()).To(ContainLines(
							fmt.Sprintf("    Running 'uv %s'", strings.Join(args, " ")),
							"      uv error stdout",
							"      uv error stderr",
						))
					})
				})
			})
		})

		context("when a lockfile exists", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(workingDir, uv.LockfileName), nil, os.ModePerm)).To(Succeed())
			})

			it.After(func() {
				Expect(os.RemoveAll(filepath.Join(workingDir, uv.LockfileName))).To(Succeed())
			})

			it("runs uv create with the cache layer available in the environment", func() {
				err := runner.Execute(uvLayerPath, uvCachePath, workingDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.CallCount).To(Equal(1))

				Expect(executions[0].Args).To(Equal([]string{
					"sync",
				}))
				venvPath := filepath.Join(uvLayerPath, "venv")
				Expect(executions[0].Env).To(ContainElement(fmt.Sprintf("HOME=%s", uvLayerPath)))
				Expect(executions[0].Env).To(ContainElement(fmt.Sprintf("HOME=%s", uvLayerPath)))
				Expect(executions[0].Env).To(ContainElement(fmt.Sprintf("VIRTUAL_ENV=%s", venvPath)))
				Expect(executions[0].Env).To(ContainElement(fmt.Sprintf("UV_PROJECT_ENVIRONMENT=%s", venvPath)))
				Expect(executions[0].Env).To(ContainElement(fmt.Sprintf("UV_WORKING_DIR=%s", workingDir)))
				Expect(executions[0].Env).To(ContainElement(fmt.Sprintf("UV_CACHE_DIR=%s", uvCachePath)))
				userFindLinks, _ := os.LookupEnv("BP_UV_FIND_LINKS")
				findLinks, _ := os.LookupEnv("UV_FIND_LINKS")
				combinedFindLinks := []string{userFindLinks, findLinks}
				Expect(executions[0].Env).To(ContainElement(fmt.Sprintf("UV_FIND_LINKS=%s", strings.TrimLeft(strings.Join(combinedFindLinks, " "), " "))))
			})

			context("failure cases", func() {
				context("when the uv env command fails to run", func() {
					it.Before(func() {
						executable.ExecuteCall.Stub = func(ex pexec.Execution) error {
							_, err := fmt.Fprintln(ex.Stdout, "uv error stdout")
							Expect(err).NotTo(HaveOccurred())
							_, err = fmt.Fprintln(ex.Stderr, "uv error stderr")
							Expect(err).NotTo(HaveOccurred())
							return errors.New("some uv failure")
						}
					})

					it("returns an error and logs the stdout and stderr output from the command", func() {
						err := runner.Execute(uvLayerPath, uvCachePath, workingDir)
						Expect(err).To(MatchError("failed to run uv command: some uv failure"))
						Expect(buffer.String()).To(ContainLines(
							"    Running 'uv sync'",
							"      uv error stdout",
							"      uv error stderr",
						))
					})
				})
			})

		})

		context("when a lockfile exists with groups specified", func() {
			it.Before(func() {
				Expect(os.Setenv("BP_UV_INSTALL_GROUPS", "dev,local")).To(Succeed())
				Expect(os.WriteFile(filepath.Join(workingDir, uv.LockfileName), nil, os.ModePerm)).To(Succeed())
			})

			it.After(func() {
				Expect(os.RemoveAll(filepath.Join(workingDir, uv.LockfileName))).To(Succeed())
				Expect(os.Unsetenv("BP_UV_INSTALL_GROUPS")).To(Succeed())
			})

			it("runs uv create with the cache layer available in the environment", func() {
				err := runner.Execute(uvLayerPath, uvCachePath, workingDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.CallCount).To(Equal(1))

				Expect(executions[0].Args).To(Equal([]string{
					"sync",
					"--group=dev",
					"--group=local",
				}))
				venvPath := filepath.Join(uvLayerPath, "venv")
				Expect(executions[0].Env).To(ContainElement(fmt.Sprintf("HOME=%s", uvLayerPath)))
				Expect(executions[0].Env).To(ContainElement(fmt.Sprintf("HOME=%s", uvLayerPath)))
				Expect(executions[0].Env).To(ContainElement(fmt.Sprintf("VIRTUAL_ENV=%s", venvPath)))
				Expect(executions[0].Env).To(ContainElement(fmt.Sprintf("UV_PROJECT_ENVIRONMENT=%s", venvPath)))
				Expect(executions[0].Env).To(ContainElement(fmt.Sprintf("UV_WORKING_DIR=%s", workingDir)))
				Expect(executions[0].Env).To(ContainElement(fmt.Sprintf("UV_CACHE_DIR=%s", uvCachePath)))
				userFindLinks, _ := os.LookupEnv("BP_UV_FIND_LINKS")
				findLinks, _ := os.LookupEnv("UV_FIND_LINKS")
				combinedFindLinks := []string{userFindLinks, findLinks}
				Expect(executions[0].Env).To(ContainElement(fmt.Sprintf("UV_FIND_LINKS=%s", strings.TrimLeft(strings.Join(combinedFindLinks, " "), " "))))
			})

			context("failure cases", func() {
				context("when the uv env command fails to run", func() {
					it.Before(func() {
						executable.ExecuteCall.Stub = func(ex pexec.Execution) error {
							_, err := fmt.Fprintln(ex.Stdout, "uv error stdout")
							Expect(err).NotTo(HaveOccurred())
							_, err = fmt.Fprintln(ex.Stderr, "uv error stderr")
							Expect(err).NotTo(HaveOccurred())
							return errors.New("some uv failure")
						}
					})

					it("returns an error and logs the stdout and stderr output from the command", func() {
						err := runner.Execute(uvLayerPath, uvCachePath, workingDir)
						Expect(err).To(MatchError("failed to run uv command: some uv failure"))
						Expect(buffer.String()).To(ContainLines(
							"    Running 'uv sync --group=dev --group=local'",
							"      uv error stdout",
							"      uv error stderr",
						))
					})
				})
			})

		})
		context("when no vendor dir or lockfile exists", func() {
			context("failure cases", func() {
				context("when no lockfile exists", func() {
					it("returns an error", func() {
						err := runner.Execute(uvLayerPath, uvCachePath, workingDir)
						Expect(err).To(MatchError(ContainSubstring("missing lock file")))
					})
				})

				context("there is an error checking for vendor directory", func() {
					it.Before(func() {
						Expect(os.Chmod(workingDir, 0000)).To(Succeed())
					})

					it.After(func() {
						Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
					})

					it("returns an error", func() {
						err := runner.Execute(uvLayerPath, uvCachePath, workingDir)
						Expect(err).To(MatchError(ContainSubstring("permission denied")))
					})
				})
			})
		})
	})
}
