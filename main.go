package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/disintegration/imaging"
)

func applyGrayscaleFilter(srcPath, destPath string) error {
	srcImage, err := imaging.Open(srcPath)
	if err != nil {
		return err
	}

	grayscaleImage := imaging.Grayscale(srcImage)

	err = imaging.Save(grayscaleImage, destPath)
	if err != nil {
		return err
	}

	return nil
}

func applyBlurFilter(srcPath, destPath string) error {
	srcImage, err := imaging.Open(srcPath)
	if err != nil {
		return err
	}

	blurredImage := imaging.Blur(srcImage, 5.0)

	err = imaging.Save(blurredImage, destPath)
	if err != nil {
		return err
	}

	return nil
}

func applyFilters(srcPath, destPath string, filter string, wg *sync.WaitGroup, ch chan string) {
	defer wg.Done()

	fileList, err := ioutil.ReadDir(srcPath)
	if err != nil {
		fmt.Printf("Error reading directory: %s\n", err.Error())
		return
	}

	for _, file := range fileList {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		srcFilePath := filepath.Join(srcPath, fileName)
		destFilePath := filepath.Join(destPath, fileName)

		switch filter {
		case "grayscale":
			err := applyGrayscaleFilter(srcFilePath, destFilePath)
			if err != nil {
				fmt.Printf("Error applying grayscale filter to %s: %s\n", fileName, err.Error())
			}
		case "blur":
			err := applyBlurFilter(srcFilePath, destFilePath)
			if err != nil {
				fmt.Printf("Error applying blur filter to %s: %s\n", fileName, err.Error())
			}
		default:
			fmt.Printf("Invalid filter: %s\n", filter)
		}

		ch <- fileName
	}
}

func processImagesWithWaitGroup(srcPath, destPath, filter string) {
	var wg sync.WaitGroup
	ch := make(chan string)

	wg.Add(1)
	go applyFilters(srcPath, destPath, filter, &wg, ch)

	go func() {
		wg.Wait()
		close(ch)
	}()

	for fileName := range ch {
		fmt.Printf("Finished processing: %s\n", fileName)
	}
}

func processImagesWithChannel(srcPath, destPath, filter string) {
	fileList, err := ioutil.ReadDir(srcPath)
	if err != nil {
		fmt.Printf("Error reading directory: %s\n", err.Error())
		return
	}

	ch := make(chan string, len(fileList))

	var wg sync.WaitGroup
	for i := 0; i < len(fileList); i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			file := fileList[index]
			if file.IsDir() {
				return
			}

			fileName := file.Name()
			srcFilePath := filepath.Join(srcPath, fileName)
			destFilePath := filepath.Join(destPath, fileName)

			switch filter {
			case "grayscale":
				err := applyGrayscaleFilter(srcFilePath, destFilePath)
				if err != nil {
					fmt.Printf("Error applying grayscale filter to %s: %s\n", fileName, err.Error())
				}
			case "blur":
				err := applyBlurFilter(srcFilePath, destFilePath)
				if err != nil {
					fmt.Printf("Error applying blur filter to %s: %s\n", fileName, err.Error())
				}
			default:
				fmt.Printf("Invalid filter: %s\n", filter)
			}

			ch <- fileName
		}(i)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for fileName := range ch {
		fmt.Printf("Finished processing: %s\n", fileName)
	}
}

func main() {
	srcPath := flag.String("src", "", "Source folder containing the images")
	destPath := flag.String("dst", "", "Destination folder to save the filtered images")
	filter := flag.String("filter", "", "Filter to apply (grayscale or blur)")
	task := flag.String("task", "", "Task method to use (waitgrp or channel)")

	flag.Parse()

	if *srcPath == "" || *destPath == "" || *filter == "" || *task == "" {
		fmt.Println("Usage: imggo -src <source_folder> -dst <destination_folder> -filter <filter_type> -task <task_method>")
		return
	}

	switch *task {
	case "waitgrp":
		processImagesWithWaitGroup(*srcPath, *destPath, *filter)
	case "channel":
		processImagesWithChannel(*srcPath, *destPath, *filter)
	default:
		fmt.Printf("Invalid task method: %s\n", *task)
	}
}
