package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// Vectorize PNG images to SVG using ImageMagick or similar tools
func vectorizeImage(inputPath, outputPath string) error {
	// Try ImageMagick's magick command first
	cmd := exec.Command("magick", inputPath,
		"-define", "png:preserve-colormap",
		"-filter", "point", // Preserve sharp pixels
		"-resize", "128x128", // Scale up 4x for crisp vectors
		outputPath)

	if err := cmd.Run(); err != nil {
		// Fallback to convert command
		cmd = exec.Command("convert", inputPath,
			"-define", "png:preserve-colormap",
			"-filter", "point",
			"-resize", "128x128",
			outputPath)

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("ImageMagick not found or failed: %v", err)
		}
	}

	return nil
}

// Create SVG wrapper for pixel-perfect scaling
func createPixelPerfectSVG(pngPath, svgPath string) error {
	// Get relative path from SVG to PNG (SVG is in svg/, PNG is in png/)
	pngRelativePath := "../png/" + filepath.Base(pngPath)

	svgContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"
     width="32" height="32" viewBox="0 0 32 32">
  <defs>
    <style>
      .pixel-art {
        image-rendering: -moz-crisp-edges;
        image-rendering: -webkit-crisp-edges;
        image-rendering: pixelated;
        image-rendering: crisp-edges;
      }
    </style>
  </defs>
  <image href="%s" class="pixel-art" width="32" height="32"/>
</svg>`, pngRelativePath)

	return os.WriteFile(svgPath, []byte(svgContent), 0644)
}

var vectorizeCmd = &cobra.Command{
	Use:   "vectorize [run_directory]",
	Short: "Convert PNG images to pixel-perfect SVG",
	Long:  "Convert generated PNG images to SVG format with pixel-perfect scaling support",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var inputDir string

		if len(args) > 0 {
			// Use specific run directory
			inputDir = filepath.Join("www/res/img/items", args[0])
		} else {
			// Default to latest run directory
			baseDir := "www/res/img/items"
			entries, err := os.ReadDir(baseDir)
			if err != nil {
				fmt.Printf("Error reading items directory: %v\n", err)
				os.Exit(1)
			}

			var latestRunDir string
			var latestTime time.Time

			for _, entry := range entries {
				if entry.IsDir() && strings.HasPrefix(entry.Name(), "run_") {
					info, err := entry.Info()
					if err != nil {
						continue
					}
					if info.ModTime().After(latestTime) {
						latestTime = info.ModTime()
						latestRunDir = entry.Name()
					}
				}
			}

			if latestRunDir == "" {
				fmt.Println("No run directories found. Please specify a run directory or generate images first.")
				os.Exit(1)
			}

			inputDir = filepath.Join(baseDir, latestRunDir)
			fmt.Printf("Using latest run directory: %s\n", latestRunDir)
		}

		outputDir := inputDir + "/svg"

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			os.Exit(1)
		}

		pngDir := filepath.Join(inputDir, "png")
		err := filepath.Walk(pngDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !strings.HasSuffix(strings.ToLower(path), ".png") {
				return nil
			}

			baseName := strings.TrimSuffix(filepath.Base(path), ".png")
			svgPath := filepath.Join(outputDir, baseName+".svg")

			fmt.Printf("Converting %s to SVG...", baseName)

			if err := createPixelPerfectSVG(path, svgPath); err != nil {
				fmt.Printf(" ERROR: %v\n", err)
				return nil
			}

			fmt.Printf(" SUCCESS\n")
			return nil
		})

		if err != nil {
			fmt.Printf("Error processing files: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nVectorization complete! SVG files saved to: %s\n", outputDir)
		fmt.Println("These SVG files will scale perfectly without blur when resized.")
	},
}

var scaleTestCmd = &cobra.Command{
	Use:   "scale-test",
	Short: "Create test images at different scales",
	Run: func(cmd *cobra.Command, args []string) {
		inputDir := "www/res/img/items/svg"
		outputDir := "www/res/img/items/scaled"
		scales := []int{64, 128, 256, 512}

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			os.Exit(1)
		}

		err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !strings.HasSuffix(strings.ToLower(path), ".svg") {
				return nil
			}

			baseName := strings.TrimSuffix(filepath.Base(path), ".svg")

			for _, scale := range scales {
				outputPath := filepath.Join(outputDir, fmt.Sprintf("%s_%dx%d.png", baseName, scale, scale))

				fmt.Printf("Scaling %s to %dx%d...", baseName, scale, scale)

				// Use ImageMagick to render SVG at specific size
				cmd := exec.Command("magick", path,
					"-resize", fmt.Sprintf("%dx%d", scale, scale),
					outputPath)

				if err := cmd.Run(); err != nil {
					fmt.Printf(" ERROR: %v\n", err)
					continue
				}

				fmt.Printf(" SUCCESS\n")
			}

			return nil
		})

		if err != nil {
			fmt.Printf("Error processing files: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nScale test complete! Files saved to: %s\n", outputDir)
	},
}

func init() {
	rootCmd.AddCommand(vectorizeCmd)
	rootCmd.AddCommand(scaleTestCmd)
}