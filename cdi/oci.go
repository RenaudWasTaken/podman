package cdi

import (
	"fmt"
	spec "github.com/opencontainers/runtime-spec/specs-go"
)

func (s *Spec) ApplyDeviceOCIEdits(config *spec.Spec, dev string) error {
	for _, d := range s.Devices {
		if d.Name != dev {
			continue
		}

		return ApplyOCIEdits(config, d.ContainerEdits)
	}

	return fmt.Errorf("CDI: device %q not found", dev)
}

func (s *Spec) ApplyOCIEdits(config *spec.Spec) error {
	return ApplyOCIEdits(config, s.ContainerEdits)
}

func ApplyOCIEdits(config *spec.Spec, edits *ContainerEdits) error {
	if edits == nil {
		return nil
	}

	for _, d := range edits.DeviceNodes {
		config.Mounts = append(config.Mounts, toOCIDevice(d))
	}

	for _, m := range edits.Mounts {
		config.Mounts = append(config.Mounts, toOCIMount(m))
	}

	for _, h := range edits.Hooks {
		switch h.HookName {
		case "prestart": config.Hooks.Prestart = append(config.Hooks.Prestart, toOCIHook(h))
		case "createRuntime": config.Hooks.CreateRuntime = append(config.Hooks.CreateRuntime, toOCIHook(h))
		case "createContainer": config.Hooks.CreateContainer = append(config.Hooks.CreateContainer, toOCIHook(h))
		case "startContainer": config.Hooks.StartContainer = append(config.Hooks.StartContainer, toOCIHook(h))
		case "poststart": config.Hooks.Poststart = append(config.Hooks.Poststart, toOCIHook(h))
		case "poststop": config.Hooks.Poststop = append(config.Hooks.Poststop, toOCIHook(h))
	default:
		fmt.Printf("CDI: Unknown hook %q\n", h.HookName)
		}
	}

	return nil
}

func toOCIHook(h *Hook) spec.Hook {
	return spec.Hook{
		Path: h.Path,
		Args: h.Args,
		Env: h.Env,
		Timeout: h.Timeout,
	}
}

func toOCIMount(m *Mount) spec.Mount {
	return spec.Mount{
		Source: m.HostPath,
		Destination: m.ContainerPath,
		Options: m.Options,
	}
}

func toOCIDevice(d *DeviceNode) spec.Mount {
	return spec.Mount{
		Source: d.HostPath,
		Destination: d.ContainerPath,
		Options: d.Permissions,
	}
}
