package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInertiaVueTemplateStructure validates Inertia-Vue templates exist.
func TestInertiaVueTemplateStructure(t *testing.T) {
	basePath := filepath.Join("..", "rituals", "blog", "templates", "frontend", "inertia")
	
	requiredFiles := []string{
		"app.js.tmpl",
		"esbuild.config.js.tmpl",
		"package.json.tmpl",
		"pages/Home.vue.tmpl",
		"pages/Posts/Index.vue.tmpl",
		"pages/Posts/Show.vue.tmpl",
		"pages/Posts/Edit.vue.tmpl",
		"components/Layout.vue.tmpl",
		"components/Header.vue.tmpl",
		"components/Footer.vue.tmpl",
	}
	
	for _, file := range requiredFiles {
		path := filepath.Join(basePath, file)
		_, err := os.Stat(path)
		assert.NoError(t, err, "Template file should exist: %s", file)
	}
}

// TestInertiaAppTemplate validates the app.js template.
func TestInertiaAppTemplate(t *testing.T) {
	path := filepath.Join("..", "rituals", "blog", "templates", "frontend", "inertia", "app.js.tmpl")
	
	content, err := os.ReadFile(path)
	require.NoError(t, err, "Should read app.js.tmpl")
	
	contentStr := string(content)
	
	// Check for required Inertia setup
	assert.Contains(t, contentStr, "createInertiaApp", "Should import createInertiaApp")
	assert.Contains(t, contentStr, "createSSRApp", "Should import createSSRApp or createApp")
	assert.Contains(t, contentStr, "resolve:", "Should have page resolver")
}

// TestEsbuildConfigTemplate validates the esbuild config template.
func TestEsbuildConfigTemplate(t *testing.T) {
	path := filepath.Join("..", "rituals", "blog", "templates", "frontend", "inertia", "esbuild.config.js.tmpl")
	
	content, err := os.ReadFile(path)
	require.NoError(t, err, "Should read esbuild.config.js.tmpl")
	
	contentStr := string(content)
	
	// Check for required esbuild configuration
	assert.Contains(t, contentStr, "entryPoints", "Should have entryPoints")
	assert.Contains(t, contentStr, "bundle", "Should enable bundle")
	assert.Contains(t, contentStr, "outdir", "Should have outdir")
	assert.Contains(t, contentStr, ".vue", "Should handle Vue files")
}

// TestPackageJsonTemplate validates the package.json template.
func TestPackageJsonTemplate(t *testing.T) {
	path := filepath.Join("..", "rituals", "blog", "templates", "frontend", "inertia", "package.json.tmpl")
	
	content, err := os.ReadFile(path)
	require.NoError(t, err, "Should read package.json.tmpl")
	
	contentStr := string(content)
	
	// Check for required dependencies
	assert.Contains(t, contentStr, "@toutaio/inertia-vue", "Should include @toutaio/inertia-vue")
	assert.Contains(t, contentStr, "vue", "Should include vue")
	assert.Contains(t, contentStr, "esbuild", "Should include esbuild")
	
	// Check for scripts
	assert.Contains(t, contentStr, "\"dev\"", "Should have dev script")
	assert.Contains(t, contentStr, "\"build\"", "Should have build script")
}

// TestVuePageTemplates validates Vue page component templates.
func TestVuePageTemplates(t *testing.T) {
	basePath := filepath.Join("..", "rituals", "blog", "templates", "frontend", "inertia", "pages")
	
	tests := []struct {
		file           string
		shouldContain  []string
	}{
		{
			file: "Home.vue.tmpl",
			shouldContain: []string{
				"<script setup>",
				"</script>",
				"<template>",
				"</template>",
			},
		},
		{
			file: "Posts/Index.vue.tmpl",
			shouldContain: []string{
				"<script setup>",
				"defineProps",
				"posts",
				"Link",
			},
		},
		{
			file: "Posts/Show.vue.tmpl",
			shouldContain: []string{
				"<script setup>",
				"defineProps",
				"post",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			path := filepath.Join(basePath, tt.file)
			content, err := os.ReadFile(path)
			require.NoError(t, err, "Should read %s", tt.file)
			
			contentStr := string(content)
			for _, expected := range tt.shouldContain {
				assert.Contains(t, contentStr, expected, "Should contain: %s", expected)
			}
		})
	}
}

// TestVueComponentTemplates validates Vue component templates.
func TestVueComponentTemplates(t *testing.T) {
	basePath := filepath.Join("..", "rituals", "blog", "templates", "frontend", "inertia", "components")
	
	tests := []struct {
		file          string
		shouldContain []string
	}{
		{
			file: "Layout.vue.tmpl",
			shouldContain: []string{
				"<script setup>",
				"<slot",
				"Header",
				"Footer",
			},
		},
		{
			file: "Header.vue.tmpl",
			shouldContain: []string{
				"<script setup>",
				"Link",
				"{{ .app_name }}",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			path := filepath.Join(basePath, tt.file)
			content, err := os.ReadFile(path)
			require.NoError(t, err, "Should read %s", tt.file)
			
			contentStr := string(content)
			for _, expected := range tt.shouldContain {
				assert.Contains(t, contentStr, expected, "Should contain: %s", expected)
			}
		})
	}
}
