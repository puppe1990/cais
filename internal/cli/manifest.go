package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

// Generated-file manifest (#169): generators record content hashes so destroy
// can warn before deleting files the user has since modified. Only whole-new
// file outputs are tracked; patched shared files (store.go, routes.go) change
// legitimately between generations and are excluded.

const generatedManifestRel = ".cais-generated.json"

func readGeneratedManifest(dir string) map[string]string {
	out := map[string]string{}
	raw, err := os.ReadFile(filepath.Join(dir, generatedManifestRel))
	if err != nil {
		return out
	}
	_ = json.Unmarshal(raw, &out)
	return out
}

func writeGeneratedManifest(dir string, entries map[string]string) error {
	raw, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, generatedManifestRel), append(raw, '\n'), 0o644)
}

// recordGeneratedFiles hashes each existing rel path into the manifest,
// merging with entries from previous generations.
func recordGeneratedFiles(dir string, rels []string) error {
	entries := readGeneratedManifest(dir)
	for _, rel := range rels {
		sum, err := hashFile(filepath.Join(dir, rel))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		entries[rel] = sum
	}
	return writeGeneratedManifest(dir, entries)
}

// dropManifestEntries removes destroyed files from the manifest.
func dropManifestEntries(dir string, rels []string) error {
	entries := readGeneratedManifest(dir)
	changed := false
	for _, rel := range rels {
		if _, ok := entries[rel]; ok {
			delete(entries, rel)
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return writeGeneratedManifest(dir, entries)
}

// fileDiffersFromManifest reports whether rel exists, is tracked, and its
// content no longer matches the recorded hash.
func fileDiffersFromManifest(dir, rel string) bool {
	want, ok := readGeneratedManifest(dir)[rel]
	if !ok || want == "" {
		return false
	}
	got, err := hashFile(filepath.Join(dir, rel))
	if err != nil {
		return false
	}
	return got != want
}

func hashFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}
