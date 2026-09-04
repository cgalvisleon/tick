package tick

import "strings"

// parseKV splits args into "key:value" tokens and bare (non key:value) words,
// in the style tick uses for project/task/status commands (e.g.
// "code:TASK001 name:Foo tag"). Order of bare words is preserved.
func parseKV(args []string) (kv map[string]string, bare []string) {
	kv = map[string]string{}
	for _, arg := range args {
		key, value, ok := strings.Cut(arg, ":")
		if !ok || key == "" {
			bare = append(bare, arg)
			continue
		}
		kv[key] = value
	}
	return kv, bare
}
