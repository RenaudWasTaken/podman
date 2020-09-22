package cdi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	spec "github.com/opencontainers/runtime-spec/specs-go"
)

const root = "/etc/cdi"
const defaultRuntime = "all"

type specMap map[string]map[string]*Spec

// loadDevices returns a map[vendor][runtime]Spec
func loadDevices() specMap {
	var files []string
	vendor := make(specMap)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() {
			fmt.Printf("CDI: Skipped path %s\n", path)
			return nil
		}

		if filepath.Ext(path) != ".json" {
			fmt.Printf("CDI: Skipped non json terminated file %s\n", path)
		}

		files = append(files, path)
		return nil
	})

	if err != nil {
		fmt.Printf("Failed to explore CDI directory\n")
		return nil
	}

	for _, path := range files {
		fmt.Printf("CDI: Reading path %s\n", path)
		spec, err := loadCDIFile(path)
		if err != nil {
			fmt.Printf("CDI: Failed to parse file %q with error %q, skipping it\n", path, err)
			continue
		}

		vndr := spec.Kind
		runtimes := spec.ContainerRuntime

		if _, ok := vendor[vndr]; !ok {
			vendor[vndr] = make(map[string]*Spec)
		}

		for _, runtime := range runtimes {
			if _, ok := vendor[vndr][runtime]; ok {
				fmt.Printf("CDI: Duplicate CDI spec for vendor %q and runtime %q at path %q, skipping it\n", vndr, runtime, path)
			}

			vendor[vndr][runtime] = spec
		}
		if len(runtimes) == 0 {
			vendor[vndr][defaultRuntime] = spec
		}
	}

	return vendor
}

func loadCDIFile(path string) (*Spec, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var spec *Spec
	err = json.Unmarshal([]byte(file), &spec)
	if err != nil {
		return nil, err
	}

	return spec, nil
}

func (m specMap) GetSpec(dev string) *Spec {
	for _, runtimes := range m {
		s, ok := runtimes[defaultRuntime]
		if !ok {
			continue
		}

		fmt.Printf("CDI: Devices are %+v\n", s.Devices)
		for _, d := range s.Devices {
			if d.Name != dev {
				continue
			}

			return s
		}
	}

	return nil
}

func HasDevice(dev string) bool {
	specs := loadDevices()
	fmt.Printf("CDI: Vendor map is %+v\n", specs)

	return specs.GetSpec(dev) != nil
}

func UpdateSpec(config *spec.Spec, devs []string) error {
	specs := loadDevices()
	uniqSpecs := make(map[*Spec]bool)

	for _, d := range devs {
		spec := specs.GetSpec(d)
		if spec == nil {
			return fmt.Errorf("Could not find dev %q", d)
		}

		uniqSpecs[spec] = true
		err := spec.ApplyDeviceOCIEdits(config, d)
		if err != nil {
			return err
		}
	}

	for s, _ := range uniqSpecs {
		s.ApplyOCIEdits(config)
	}

	return nil
}
