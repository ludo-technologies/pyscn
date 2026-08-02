package app

import "path/filepath"

// analysisPathIndex owns the relationship between an internal file identity
// and the first caller-facing spelling of that path.
type analysisPathIndex struct {
	reportedByIdentity map[string]string
}

func prepareAnalysisPaths(paths []string) ([]string, analysisPathIndex, error) {
	unique := make([]string, 0, len(paths))
	index := analysisPathIndex{reportedByIdentity: make(map[string]string, len(paths))}
	for _, path := range paths {
		identity, err := analysisPathIdentity(path)
		if err != nil {
			return nil, analysisPathIndex{}, err
		}
		if _, duplicate := index.reportedByIdentity[identity]; duplicate {
			continue
		}
		index.reportedByIdentity[identity] = path
		unique = append(unique, path)
	}
	return unique, index, nil
}

func (i analysisPathIndex) reportedPath(path string) (string, error) {
	identity, err := analysisPathIdentity(path)
	if err != nil {
		return "", err
	}
	if reported, ok := i.reportedByIdentity[identity]; ok {
		return reported, nil
	}
	return path, nil
}

func analysisPathIdentity(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(absolute), nil
}

// configDiscoveryTarget returns the path config discovery starts from for a
// request. The first analyzed path stands for the whole request: discovery has
// to begin inside the analyzed tree rather than at the working directory, which
// may belong to an entirely different project (issue #666).
func configDiscoveryTarget(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}
