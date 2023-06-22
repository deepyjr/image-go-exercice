package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/disintegration/imaging"
)

// applyGrayscaleFilter applique le filtre de grayscale à une image donnée.
func applyGrayscaleFilter(srcPath, destPath string) error {
	// Ouvrir l'image source
	srcImage, err := imaging.Open(srcPath)
	if err != nil {
		return err
	}

	// Appliquer le filtre de grayscale à l'image
	grayscaleImage := imaging.Grayscale(srcImage)

	// Sauvegarder l'image filtrée dans le chemin de destination
	err = imaging.Save(grayscaleImage, destPath)
	if err != nil {
		return err
	}

	return nil
}

// applyBlurFilter applique le filtre de blur à une image donnée.
func applyBlurFilter(srcPath, destPath string) error {
	// Ouvrir l'image source
	srcImage, err := imaging.Open(srcPath)
	if err != nil {
		return err
	}

	// Appliquer le filtre de blur à l'image
	blurredImage := imaging.Blur(srcImage, 5.0)

	// Sauvegarder l'image filtrée dans le chemin de destination
	err = imaging.Save(blurredImage, destPath)
	if err != nil {
		return err
	}

	return nil
}

// applyFilters parcourt le dossier source, applique les filtres spécifiés aux images et sauvegarde les images filtrées dans le dossier de destination.
// Cette fonction est utilisée avec la méthode WaitGroup pour répartir les tâches en parallèle.
func applyFilters(srcPath, destPath string, filter string, wg *sync.WaitGroup, ch chan string) {
	// Marquer la fin de la tâche lorsque la fonction se termine
	defer wg.Done()

	// Lire la liste des fichiers du dossier source
	fileList, err := ioutil.ReadDir(srcPath)
	if err != nil {
		fmt.Printf("Error reading directory: %s\n", err.Error())
		return
	}

	// Parcourir chaque fichier du dossier source
	for _, file := range fileList {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		srcFilePath := filepath.Join(srcPath, fileName)
		destFilePath := filepath.Join(destPath, fileName)

		// Appliquer le filtre spécifié à l'image
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

		// Envoyer le nom du fichier traité via le canal pour afficher une notification
		ch <- fileName
	}
}

// processImagesWithWaitGroup génère la liste des fichiers à traiter et dispatche les tâches de filtrage sur les images en utilisant la méthode WaitGroup.
func processImagesWithWaitGroup(srcPath, destPath, filter string) {
	var wg sync.WaitGroup
	ch := make(chan string)

	// Ajouter une tâche au WaitGroup pour la fonction applyFilters
	wg.Add(1)
	go applyFilters(srcPath, destPath, filter, &wg, ch)

	// Attendre la fin de toutes les tâches en utilisant le WaitGroup
	go func() {
		wg.Wait()
		close(ch)
	}()

	// Lire les messages du canal pour afficher les fichiers traités
	for fileName := range ch {
		fmt.Printf("Finished processing: %s\n", fileName)
	}
}

// processImagesWithChannel génère la liste des fichiers à traiter et dispatche les tâches de filtrage sur les images en utilisant des canaux.
func processImagesWithChannel(srcPath, destPath, filter string) {
	// Lire la liste des fichiers du dossier source
	fileList, err := ioutil.ReadDir(srcPath)
	if err != nil {
		fmt.Printf("Error reading directory: %s\n", err.Error())
		return
	}

	// Créer un canal avec une taille équivalente au nombre de fichiers pour éviter les blocages
	ch := make(chan string, len(fileList))

	var wg sync.WaitGroup
	for i := 0; i < len(fileList); i++ {
		wg.Add(1)
		go func(index int) {
			// Marquer la fin de la tâche lorsque la fonction se termine
			defer wg.Done()

			file := fileList[index]
			if file.IsDir() {
				return
			}

			fileName := file.Name()
			srcFilePath := filepath.Join(srcPath, fileName)
			destFilePath := filepath.Join(destPath, fileName)

			// Appliquer le filtre spécifié à l'image
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

			// Envoyer le nom du fichier traité via le canal pour afficher une notification
			ch <- fileName
		}(i)
	}

	// Attendre la fin de toutes les tâches en utilisant le WaitGroup
	go func() {
		wg.Wait()
		close(ch)
	}()

	// Lire les messages du canal pour afficher les fichiers traités
	for fileName := range ch {
		fmt.Printf("Finished processing: %s\n", fileName)
	}
}

func main() {
	// Analyse des arguments de ligne de commande
	srcPath := flag.String("src", "", "Source folder containing the images")
	destPath := flag.String("dst", "", "Destination folder to save the filtered images")
	filter := flag.String("filter", "", "Filter to apply (grayscale or blur)")
	task := flag.String("task", "", "Task method to use (waitgrp or channel)")

	flag.Parse()

	// Vérification des arguments requis
	if *srcPath == "" || *destPath == "" || *filter == "" || *task == "" {
		fmt.Println("Usage: imggo -src <source_folder> -dst <destination_folder> -filter <filter_type> -task <task_method>")
		return
	}

	// Exécution de la méthode de traitement appropriée en fonction du paramètre de tâche
	switch *task {
	case "waitgrp":
		processImagesWithWaitGroup(*srcPath, *destPath, *filter)
	case "channel":
		processImagesWithChannel(*srcPath, *destPath, *filter)
	default:
		fmt.Printf("Invalid task method: %s\n", *task)
	}
}
