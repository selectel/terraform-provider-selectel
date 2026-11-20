package mks

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	mksv2 "github.com/selectel/mks-go/v2/pkg"
	"github.com/selectel/mks-go/v2/pkg/kubeversion"
)

func errKubeVersionIsInvalidFmt(kubeVersion string, errMsg string) error {
	return fmt.Errorf("kube version %s is invalid: %s", kubeVersion, errMsg)
}

// kubeVersionToMajor returns the major part of a Kubernetes version.
func kubeVersionToMajor(kubeVersion string) (int, error) {
	kubeVersion = strings.TrimPrefix(kubeVersion, "v")
	kubeVersionParts := strings.Split(kubeVersion, ".")
	major, err := strconv.Atoi(kubeVersionParts[0])
	if err != nil {
		return 0, errKubeVersionIsInvalidFmt(kubeVersion, "major part is not an integer number")
	}
	if major < 0 {
		return 0, errKubeVersionIsInvalidFmt(kubeVersion, "major part is a negative number")
	}

	return major, nil
}

// kubeVersionToMinor returns the minor part of a Kubernetes version.
func kubeVersionToMinor(kubeVersion string) (int, error) {
	kubeVersion = strings.TrimPrefix(kubeVersion, "v")
	kubeVersionParts := strings.Split(kubeVersion, ".")
	if len(kubeVersionParts) < 2 {
		return 0, errKubeVersionIsInvalidFmt(kubeVersion, "expected to have major and minor version parts")
	}
	minor, err := strconv.Atoi(kubeVersionParts[1])
	if err != nil {
		return 0, errKubeVersionIsInvalidFmt(kubeVersion, "minor part is not an integer number")
	}
	if minor < 0 {
		return 0, errKubeVersionIsInvalidFmt(kubeVersion, "minor part is a negative number")
	}

	return minor, nil
}

// kubeVersionToPatch returns the patch part of a Kubernetes version.
func kubeVersionToPatch(kubeVersion string) (int, error) {
	kubeVersion = strings.TrimPrefix(kubeVersion, "v")
	kubeVersionParts := strings.Split(kubeVersion, ".")
	if len(kubeVersionParts) < 3 {
		return 0, errKubeVersionIsInvalidFmt(kubeVersion, "expected to have major, minor and patch version parts")
	}
	patch, err := strconv.Atoi(kubeVersionParts[2])
	if err != nil {
		return 0, errKubeVersionIsInvalidFmt(kubeVersion, "patch part is not an integer number")
	}
	if patch < 0 {
		return 0, errKubeVersionIsInvalidFmt(kubeVersion, "patch part is a negative number")
	}

	return patch, nil
}

// compareTwoKubeVersionsByPatch parses two Kubernetes versions, compares
// their patch versions and returns the latest one.
// It doesn't check minor version so it will give a bad result for different
// minor versions.
func compareTwoKubeVersionsByPatch(a, b string) (string, error) {
	aPatch, err := kubeVersionToPatch(a)
	if err != nil {
		return "", fmt.Errorf("unable to compare kube versions: %s", err)
	}
	bPatch, err := kubeVersionToPatch(b)
	if err != nil {
		return "", fmt.Errorf("unable to compare kube versions: %s", err)
	}
	if aPatch > bPatch {
		return a, nil
	}

	return b, nil
}

// kubeVersionTrimToMinor returns the version trimmed to its minor part, "x.y".
func kubeVersionTrimToMinor(kubeVersion string) (string, error) {
	major, err := kubeVersionToMajor(kubeVersion)
	if err != nil {
		return "", err
	}
	minor, err := kubeVersionToMinor(kubeVersion)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d.%d", major, minor), nil
}

// kubeVersionTrimToMinorIncremented returns the version trimmed to the next
// minor version, "x.(y+1)".
func kubeVersionTrimToMinorIncremented(kubeVersion string) (string, error) {
	major, err := kubeVersionToMajor(kubeVersion)
	if err != nil {
		return "", err
	}
	minor, err := kubeVersionToMinor(kubeVersion)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d.%d", major, minor+1), nil
}

// latestKubePatchVersions returns the latest patch version of every minor
// version the API serves.
func latestKubePatchVersions(ctx context.Context, client *mksv2.ServiceClient) (map[string]string, error) {
	kubeVersions, err := kubeversion.List(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("error getting kube versions: %w", err)
	}

	result := map[string]string{}
	for _, version := range kubeVersions {
		minor, err := kubeVersionTrimToMinor(valueOrZero(version.Version))
		if err != nil {
			return nil, err
		}
		current, ok := result[minor]
		if !ok {
			result[minor] = valueOrZero(version.Version)
			continue
		}
		latest, err := compareTwoKubeVersionsByPatch(valueOrZero(version.Version), current)
		if err != nil {
			return nil, err
		}
		result[minor] = latest
	}

	return result, nil
}

// The verbatim provider error texts the package cannot import.
func errCreatingObject(object string, err error) error {
	return fmt.Errorf("error creating %s: %s", object, err)
}

func errUpdatingObject(object, id string, err error) error {
	return fmt.Errorf("error updating %s '%s': %s", object, id, err)
}

func errGettingObject(object, id string, err error) error {
	return fmt.Errorf("error getting %s '%s': %s", object, id, err)
}

func errDeletingObject(object, id string, err error) error {
	return fmt.Errorf("error deleting %s '%s': %s", object, id, err)
}

const objectCluster = "cluster"
