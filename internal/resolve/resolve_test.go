package resolve

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestResolveFilenames2(t *testing.T) {
	// Create a temporary test directory structure
	tempDir, err := os.MkdirTemp("", "test_resolveFilenames2")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir) // Clean up after test

	testFile1 := filepath.Join(tempDir, "file1.txt")
	testFile2 := filepath.Join(tempDir, "file2.yaml")
	testSubDir := filepath.Join(tempDir, "subdir")
	testFile3 := filepath.Join(testSubDir, "file3.json")
	testFile4 := filepath.Join(testSubDir, "file4.xml")

	// Create test files and directories
	if err := os.WriteFile(testFile1, []byte("content1"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testFile2, []byte("content2"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(testSubDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testFile3, []byte("content3"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testFile4, []byte("content4"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Define test cases
	tests := []struct {
		name        string
		path        string
		recursive   bool
		expected    []string
		expectError bool
	}{
		{
			name:        "Single file valid extension",
			path:        testFile2,
			recursive:   false,
			expected:    []string{testFile2},
			expectError: false,
		},
		{
			name:        "Glob pattern valid extensions",
			path:        filepath.Join(tempDir, "*.yaml"),
			recursive:   false,
			expected:    []string{testFile2},
			expectError: false,
		},
		{
			name:        "Directory non-recursive",
			path:        tempDir,
			recursive:   false,
			expected:    []string{testFile2},
			expectError: false,
		},
		{
			name:        "Directory recursive",
			path:        tempDir,
			recursive:   true,
			expected:    []string{testFile2, testFile3},
			expectError: false,
		},
		{
			name:        "URL handling",
			path:        "http://example.com/file.yaml",
			recursive:   false,
			expected:    []string{"http://example.com/file.yaml"},
			expectError: false,
		},
		{
			name:        "Nonexistent file",
			path:        filepath.Join(tempDir, "nonexistent.txt"),
			recursive:   false,
			expected:    nil,
			expectError: true,
		},
	}

	// Execute test cases
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := resolveFilenamesForPatterns(test.path, test.recursive)
			if test.expectError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("did not expect an error but got: %v", err)
				}
				if !reflect.DeepEqual(result, test.expected) {
					t.Errorf("expected %v, got %v", test.expected, result)
				}
			}
		})
	}
}

func TestIsURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid HTTP URL", "http://example.com", true},
		{"Valid HTTPS URL", "https://example.com", true},
		{"Invalid URL", "example.com", false},
		{"Empty String", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isURL(tt.input)
			if result != tt.expected {
				t.Errorf("isURL(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIgnoreFile(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		extensions []string
		expected   bool
	}{
		{"Allowed Extension", "file.yaml", []string{".json", ".yaml"}, false},
		{"Disallowed Extension", "file.txt", []string{".json", ".yaml"}, true},
		{"Empty Extensions", "file.yaml", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ignoreFile(tt.path, tt.extensions)
			if result != tt.expected {
				t.Errorf("ignoreFile(%q, %v) = %v, expected %v", tt.path, tt.extensions, result, tt.expected)
			}
		})
	}
}

func TestResolveAllFiles(t *testing.T) {
	// Setup temporary files and directories for testing
	tempDir := t.TempDir()

	// Create test files
	file1 := filepath.Join(tempDir, "file1.yaml")
	file2 := filepath.Join(tempDir, "file2.json")
	file3 := filepath.Join(tempDir, "ignore.txt")
	subDir := filepath.Join(tempDir, "subdir")
	subFile := filepath.Join(subDir, "file3.yaml")

	if err := os.WriteFile(file1, []byte("test content"), 0o600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("test content"), 0o600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(file3, []byte("test content"), 0o600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatalf("Failed to create test subdir: %v", err)
	}
	if err := os.WriteFile(subFile, []byte("test content"), 0o600); err != nil {
		t.Fatalf("Failed to create test subfile: %v", err)
	}

	tests := []struct {
		name      string
		filenames []string
		recursive bool
		want      []string
		wantErr   bool
	}{
		{
			name:      "Single file",
			filenames: []string{file1},
			recursive: false,
			want:      []string{file1},
			wantErr:   false,
		},
		{
			name:      "Glob pattern",
			filenames: []string{filepath.Join(tempDir, "*.yaml")},
			recursive: false,
			want:      []string{file1},
			wantErr:   false,
		},
		{
			name:      "Directory, non-recursive",
			filenames: []string{tempDir},
			recursive: false,
			want:      []string{file1, file2},
			wantErr:   false,
		},
		{
			name:      "Directory, recursive",
			filenames: []string{tempDir},
			recursive: true,
			want:      []string{file1, file2, subFile},
			wantErr:   false,
		},
		{
			name:      "If file passed explicitly, don't check its extension",
			filenames: []string{file3},
			recursive: false,
			want:      []string{file3},
			wantErr:   false,
		},
		{
			name:      "URL as input",
			filenames: []string{"https://example.com/file.yaml"},
			recursive: false,
			want:      []string{"https://example.com/file.yaml"},
			wantErr:   false,
		},
		{
			name:      "Invalid path",
			filenames: []string{"/invalid/path/file.yaml"},
			recursive: false,
			want:      nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveAllFiles(tt.filenames, tt.recursive)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveAllFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResolveAllFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadRemoteFileContent(t *testing.T) {
	tests := []struct {
		name         string
		inputURL     string
		mockResponse string
		statusCode   int
		wantError    bool
	}{
		{
			name:         "Valid URL and HTTP response",
			inputURL:     "http://example.com/test",
			mockResponse: "file content",
			statusCode:   http.StatusOK,
			wantError:    false,
		},
		{
			name:      "Invalid URL - malformed",
			inputURL:  ":invalid-url",
			wantError: true,
		},
		{
			name:      "Invalid URL - missing host",
			inputURL:  "http:///test",
			wantError: true,
		},
		{
			name:       "Non-200 HTTP status code",
			inputURL:   "http://example.com/notfound",
			statusCode: http.StatusNotFound,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statusCode)
				fmt.Fprint(w, tt.mockResponse)
			}))
			defer server.Close()

			// Replace the input URL with the mock server's URL if statusCode is set
			inputURL := tt.inputURL
			if tt.statusCode != 0 {
				inputURL = server.URL
			}

			// Call the function under test
			got, err := readRemoteFileContent(inputURL)

			// Check for errors
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error but got one: %v", err)
				}
				// Verify the response content
				if string(got) != tt.mockResponse {
					t.Errorf("Expected response %q, got %q", tt.mockResponse, got)
				}
			}
		})
	}
}
