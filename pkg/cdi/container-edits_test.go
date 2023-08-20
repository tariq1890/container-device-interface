/*
   Copyright © 2021 The CDI Authors

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

package cdi

import (
	"testing"

	oci "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/require"

	"github.com/container-orchestrated-devices/container-device-interface/specs-go"
	cdi "github.com/container-orchestrated-devices/container-device-interface/specs-go"
)

func TestValidateContainerEdits(t *testing.T) {
	type testCase struct {
		name    string
		edits   *cdi.ContainerEdits
		invalid bool
	}
	for _, tc := range []*testCase{
		{
			name:  "valid, empty edits",
			edits: nil,
		},
		{
			name: "valid, env var",
			edits: &cdi.ContainerEdits{
				Env: []string{"BAR=BARVALUE1"},
			},
		},
		{
			name: "invalid env, empty var",
			edits: &cdi.ContainerEdits{
				Env: []string{""},
			},
			invalid: true,
		},
		{
			name: "invalid env, no var name",
			edits: &cdi.ContainerEdits{
				Env: []string{"=foo"},
			},
			invalid: true,
		},
		{
			name: "invalid env, no assignment",
			edits: &cdi.ContainerEdits{
				Env: []string{"FOOBAR"},
			},
			invalid: true,
		},
		{
			name: "valid device, path only",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path: "/dev/null",
					},
				},
			},
		},
		{
			name: "valid device, path+type",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path: "/dev/null",
						Type: "c",
					},
				},
			},
		},
		{
			name: "valid device, path+type+permissions",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path:        "/dev/null",
						Type:        "b",
						Permissions: "rwm",
					},
				},
			},
		},
		{
			name: "invalid device, empty path",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path: "",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid device, wrong type",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path: "/dev/vendorctl",
						Type: "f",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid device, wrong permissions",
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path:        "/dev/vendorctl",
						Type:        "b",
						Permissions: "to land",
					},
				},
			},
			invalid: true,
		},
		{
			name: "valid mount",
			edits: &cdi.ContainerEdits{
				Mounts: []*cdi.Mount{
					{
						HostPath:      "/dev/vendorctl",
						ContainerPath: "/dev/vendorctl",
					},
				},
			},
		},
		{
			name: "invalid mount, empty host path",
			edits: &cdi.ContainerEdits{
				Mounts: []*cdi.Mount{
					{
						HostPath:      "",
						ContainerPath: "/dev/vendorctl",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid mount, empty container path",
			edits: &cdi.ContainerEdits{
				Mounts: []*cdi.Mount{
					{
						HostPath:      "/dev/vendorctl",
						ContainerPath: "",
					},
				},
			},
			invalid: true,
		},
		{
			name: "valid hooks",
			edits: &cdi.ContainerEdits{
				Hooks: []*cdi.Hook{
					{
						HookName: "prestart",
						Path:     "/usr/local/bin/prestart-vendor-hook",
						Args:     []string{"--verbose"},
						Env:      []string{"VENDOR_ENV1=value1"},
					},
					{
						HookName: "createRuntime",
						Path:     "/usr/local/bin/cr-vendor-hook",
						Args:     []string{"--debug"},
						Env:      []string{"VENDOR_ENV2=value2"},
					},
					{
						HookName: "createContainer",
						Path:     "/usr/local/bin/cc-vendor-hook",
						Args:     []string{"--create"},
						Env:      []string{"VENDOR_ENV3=value3"},
					},
					{
						HookName: "startContainer",
						Path:     "/usr/local/bin/sc-vendor-hook",
						Args:     []string{"--start"},
						Env:      []string{"VENDOR_ENV4=value4"},
					},
					{
						HookName: "poststart",
						Path:     "/usr/local/bin/poststart-vendor-hook",
						Env:      []string{"VENDOR_ENV5=value5"},
					},
					{
						HookName: "poststop",
						Path:     "/usr/local/bin/poststop-vendor-hook",
					},
				},
			},
		},
		{
			name: "invalid hook, empty path",
			edits: &cdi.ContainerEdits{
				Hooks: []*cdi.Hook{
					{
						HookName: "prestart",
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid hook, wrong hook name",
			edits: &cdi.ContainerEdits{
				Hooks: []*cdi.Hook{
					{
						HookName: "misCreateRuntime",
						Path:     "/usr/local/bin/cr-vendor-hook",
						Args:     []string{"--debug"},
						Env:      []string{"VENDOR_ENV2=value2"},
					},
				},
			},
			invalid: true,
		},
		{
			name: "invalid hook, wrong env",
			edits: &cdi.ContainerEdits{
				Hooks: []*cdi.Hook{
					{
						HookName: "poststart",
						Path:     "/usr/local/bin/cr-vendor-hook",
						Args:     []string{"--debug"},
						Env:      []string{"=value2"},
					},
				},
			},
			invalid: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			edits := ContainerEdits{tc.edits}
			err := edits.Validate()
			if tc.invalid {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestApplyContainerEdits(t *testing.T) {
	type testCase struct {
		name   string
		spec   *oci.Spec
		edits  *cdi.ContainerEdits
		result *oci.Spec
	}
	for _, tc := range []*testCase{
		{
			name:   "empty spec, empty edits",
			spec:   &oci.Spec{},
			edits:  nil,
			result: &oci.Spec{},
		},
		{
			name: "empty spec, env var",
			spec: &oci.Spec{},
			edits: &cdi.ContainerEdits{
				Env: []string{"BAR=BARVALUE1"},
			},
			result: &oci.Spec{
				Process: &oci.Process{
					Env: []string{
						"BAR=BARVALUE1",
					},
				},
			},
		},
		{
			name: "empty spec, device",
			spec: &oci.Spec{},
			edits: &cdi.ContainerEdits{
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path: "/dev/null",
					},
				},
			},
			result: &oci.Spec{
				Linux: &oci.Linux{
					Devices: []oci.LinuxDevice{
						{
							Path:  "/dev/null",
							Type:  "c",
							Major: 1,
							Minor: 3,
						},
					},
					Resources: &oci.LinuxResources{
						Devices: []oci.LinuxDeviceCgroup{
							{
								Allow:  true,
								Type:   "c",
								Major:  int64ptr(1),
								Minor:  int64ptr(3),
								Access: "rwm",
							},
						},
					},
				},
			},
		},
		{
			name: "empty spec, device, env var",
			spec: &oci.Spec{},
			edits: &cdi.ContainerEdits{
				Env: []string{
					"FOO=BAR",
				},
				DeviceNodes: []*cdi.DeviceNode{
					{
						Path: "/dev/null",
						Type: "c",
					},
				},
			},
			result: &oci.Spec{
				Process: &oci.Process{
					Env: []string{
						"FOO=BAR",
					},
				},
				Linux: &oci.Linux{
					Devices: []oci.LinuxDevice{
						{
							Path:  "/dev/null",
							Type:  "c",
							Major: 1,
							Minor: 3,
						},
					},
					Resources: &oci.LinuxResources{
						Devices: []oci.LinuxDeviceCgroup{
							{
								Allow:  true,
								Type:   "c",
								Major:  int64ptr(1),
								Minor:  int64ptr(3),
								Access: "rwm",
							},
						},
					},
				},
			},
		},
		{
			name: "empty spec, mount",
			spec: &oci.Spec{},
			edits: &cdi.ContainerEdits{
				Mounts: []*cdi.Mount{
					{
						HostPath:      "/dev/host-vendorctl",
						ContainerPath: "/dev/cntr-vendorctl",
					},
				},
			},
			result: &oci.Spec{
				Mounts: []oci.Mount{
					{
						Source:      "/dev/host-vendorctl",
						Destination: "/dev/cntr-vendorctl",
					},
				},
			},
		},
		{
			name: "empty spec, hooks",
			spec: &oci.Spec{},
			edits: &cdi.ContainerEdits{
				Hooks: []*cdi.Hook{
					{
						HookName: "prestart",
						Path:     "/usr/local/bin/prestart-vendor-hook",
						Args:     []string{"--verbose"},
						Env:      []string{"VENDOR_ENV1=value1"},
					},
					{
						HookName: "createRuntime",
						Path:     "/usr/local/bin/cr-vendor-hook",
						Args:     []string{"--debug"},
						Env:      []string{"VENDOR_ENV2=value2"},
					},
					{
						HookName: "createContainer",
						Path:     "/usr/local/bin/cc-vendor-hook",
						Args:     []string{"--create"},
						Env:      []string{"VENDOR_ENV3=value3"},
					},
					{
						HookName: "startContainer",
						Path:     "/usr/local/bin/sc-vendor-hook",
						Args:     []string{"--start"},
						Env:      []string{"VENDOR_ENV4=value4"},
					},
					{
						HookName: "poststart",
						Path:     "/usr/local/bin/poststart-vendor-hook",
						Env:      []string{"VENDOR_ENV5=value5"},
					},
					{
						HookName: "poststop",
						Path:     "/usr/local/bin/poststop-vendor-hook",
					},
				},
			},
			result: &oci.Spec{
				Hooks: &oci.Hooks{
					Prestart: []oci.Hook{
						{
							Path: "/usr/local/bin/prestart-vendor-hook",
							Args: []string{"--verbose"},
							Env:  []string{"VENDOR_ENV1=value1"},
						},
					},
					CreateRuntime: []oci.Hook{
						{
							Path: "/usr/local/bin/cr-vendor-hook",
							Args: []string{"--debug"},
							Env:  []string{"VENDOR_ENV2=value2"},
						},
					},
					CreateContainer: []oci.Hook{
						{
							Path: "/usr/local/bin/cc-vendor-hook",
							Args: []string{"--create"},
							Env:  []string{"VENDOR_ENV3=value3"},
						},
					},
					StartContainer: []oci.Hook{
						{
							Path: "/usr/local/bin/sc-vendor-hook",
							Args: []string{"--start"},
							Env:  []string{"VENDOR_ENV4=value4"},
						},
					},
					Poststart: []oci.Hook{
						{
							Path: "/usr/local/bin/poststart-vendor-hook",
							Env:  []string{"VENDOR_ENV5=value5"},
						},
					},
					Poststop: []oci.Hook{
						{
							Path: "/usr/local/bin/poststop-vendor-hook",
						},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			edits := ContainerEdits{tc.edits}
			err := edits.Validate()
			require.NoError(t, err)
			err = edits.Apply(tc.spec)
			require.NoError(t, err)
			require.Equal(t, tc.result, tc.spec)
		})
	}
}

func TestAppend(t *testing.T) {
	type testCase struct {
		name   string
		dst    *ContainerEdits
		src    []*ContainerEdits
		result *ContainerEdits
	}
	for _, tc := range []*testCase{
		{
			name: "merge nil into nil",
			dst:  nil,
			src: []*ContainerEdits{
				nil,
			},
			result: nil,
		},
		{
			name: "merge non-nil into nil",
			dst:  nil,
			src: []*ContainerEdits{
				{
					ContainerEdits: &specs.ContainerEdits{
						Env: []string{
							"var1=val1",
						},
					},
				},
			},
			result: &ContainerEdits{
				ContainerEdits: &specs.ContainerEdits{
					Env: []string{
						"var1=val1",
					},
				},
			},
		},
		{
			name: "merge nil into non-nil",
			dst: &ContainerEdits{
				ContainerEdits: &specs.ContainerEdits{
					Env: []string{
						"var1=val1",
					},
				},
			},
			src: []*ContainerEdits{
				nil,
			},
			result: &ContainerEdits{
				ContainerEdits: &specs.ContainerEdits{
					Env: []string{
						"var1=val1",
					},
				},
			},
		},
		{
			name: "merge multiple into non-nil",
			dst: &ContainerEdits{
				ContainerEdits: &specs.ContainerEdits{
					Env: []string{
						"var0=val0",
					},
				},
			},
			src: []*ContainerEdits{
				{
					ContainerEdits: &specs.ContainerEdits{
						Env: []string{
							"var1=val1",
						},
						DeviceNodes: []*specs.DeviceNode{
							{
								Path: "/dev/dev1",
							},
						},
					},
				},
				{
					ContainerEdits: &specs.ContainerEdits{
						Env: []string{
							"var2=val2",
							"var3=val3",
						},
						DeviceNodes: []*specs.DeviceNode{
							{
								Path: "/dev/dev2",
							},
							{
								Path: "/dev/dev3",
							},
						},
					},
				},
				{
					ContainerEdits: &specs.ContainerEdits{
						Env: []string{
							"var4=val4",
						},
						DeviceNodes: []*specs.DeviceNode{
							{
								Path: "/dev/dev4",
							},
						},
					},
				},
			},
			result: &ContainerEdits{
				ContainerEdits: &specs.ContainerEdits{
					Env: []string{
						"var0=val0",
						"var1=val1",
						"var2=val2",
						"var3=val3",
						"var4=val4",
					},
					DeviceNodes: []*specs.DeviceNode{
						{
							Path: "/dev/dev1",
						},
						{
							Path: "/dev/dev2",
						},
						{
							Path: "/dev/dev3",
						},
						{
							Path: "/dev/dev4",
						},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dst := tc.dst
			for _, src := range tc.src {
				dst = dst.Append(src)
			}
			require.Equal(t, tc.result, dst, "append container edits")
		})
	}
}
