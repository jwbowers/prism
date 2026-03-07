package templates

import (
	"context"
	"log"
	"time"
)

// communityTemplateTTL is the time-to-live for cached community templates.
// Matches DefaultCacheOptions().TTL (1 hour).
const communityTemplateTTL = 1 * time.Hour

// FetchCommunityTemplates loads templates from all enabled community sources.
//
// Results are served from the local TTL cache when fresh (keyed on
// source.LastSync). Pass force=true to skip the cache and always fetch from
// the remote registry.
//
// Errors per-source are non-fatal: if a source is unreachable, stale cached
// templates are used as a fallback. Other sources continue to be processed.
// The returned slice may be empty but is never nil on a nil error.
func FetchCommunityTemplates(force bool) ([]*Template, error) {
	sm := NewSourceManager()

	// On first run (no sources configured), auto-register the official registry.
	if len(sm.List()) == 0 {
		if err := sm.AddDefaultSources(); err != nil {
			log.Printf("[COMMUNITY TEMPLATES] Warning: could not add default sources: %v", err)
		}
	}

	cache := NewCommunityCache(DefaultCacheOptions())
	adapter := NewGitHubAdapter()

	var allTemplates []*Template

	for _, source := range sm.List() {
		if !source.Enabled {
			continue
		}

		// Decide whether the local cache is still fresh.
		useCache := !force &&
			!source.LastSync.IsZero() &&
			time.Since(source.LastSync) <= communityTemplateTTL

		if useCache {
			cached := cachedSourceTemplates(cache, source.URL, source.Branch)
			if len(cached) > 0 {
				allTemplates = append(allTemplates, cached...)
				continue
			}
			// Cache is empty or all entries expired — fall through to fetch.
		}

		// Fetch from the remote source (30-second timeout).
		fetchCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		fetched, err := adapter.FetchTemplates(fetchCtx, source.URL, source.Branch)
		cancel()

		if err != nil {
			log.Printf("[COMMUNITY TEMPLATES] Warning: could not fetch from %s: %v", source.URL, err)
			// Serve stale cache as a graceful fallback.
			stale := cachedSourceTemplates(cache, source.URL, source.Branch)
			if len(stale) > 0 {
				log.Printf("[COMMUNITY TEMPLATES] Using %d stale cached templates from %s",
					len(stale), source.Name)
				allTemplates = append(allTemplates, stale...)
			}
			continue
		}

		// Refresh the on-disk cache for this source.
		_ = cache.InvalidateSource(source.URL, source.Branch)
		for _, tmpl := range fetched {
			_ = cache.Set(source.URL, source.Branch, communityTemplateCachePath(tmpl), tmpl)
		}
		_ = sm.UpdateLastSync(source.Name)

		log.Printf("[COMMUNITY TEMPLATES] Fetched %d templates from %s", len(fetched), source.Name)
		allTemplates = append(allTemplates, fetched...)
	}

	return allTemplates, nil
}

// LoadCommunityTemplatesIntoRegistry injects community templates into an
// already-scanned TemplateRegistry. Core / local templates always take
// precedence — community templates are only added when no local template
// with the same name exists. Errors are logged and never returned.
func LoadCommunityTemplatesIntoRegistry(registry *TemplateRegistry) {
	tmpl, err := FetchCommunityTemplates(false)
	if err != nil {
		log.Printf("[COMMUNITY TEMPLATES] Warning: %v", err)
		return
	}

	added := 0
	for _, t := range tmpl {
		if t.Name == "" {
			continue
		}
		// Don't override core / local templates.
		if _, exists := registry.Templates[t.Name]; exists {
			continue
		}
		registry.Templates[t.Name] = t
		if t.Slug != "" {
			registry.SlugIndex[t.Slug] = t.Name
		}
		added++
	}

	if added > 0 {
		log.Printf("[COMMUNITY TEMPLATES] Added %d community templates to registry", added)
	}
}

// cachedSourceTemplates returns all non-expired cached templates for a source.
func cachedSourceTemplates(cache *CommunityCache, sourceURL, branch string) []*Template {
	entries, err := cache.ListCachedTemplates()
	if err != nil {
		return nil
	}
	var result []*Template
	for _, entry := range entries {
		if entry.SourceURL != sourceURL || entry.SourceBranch != branch {
			continue
		}
		if time.Since(entry.CachedAt) > communityTemplateTTL {
			continue
		}
		result = append(result, entry.Template)
	}
	return result
}

// communityTemplateCachePath returns a stable cache path for a template.
func communityTemplateCachePath(t *Template) string {
	if t.Slug != "" {
		return "templates/" + t.Slug + ".yml"
	}
	return "templates/" + t.Name + ".yml"
}
