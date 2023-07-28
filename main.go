package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.POST("/unidock", func(c *gin.Context) {
		receptor := c.PostForm("receptor")
		ligand := c.PostForm("ligand")

		root_path, err := os.Getwd()
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		err = os.Mkdir(path.Join(root_path, "tmp"), 0755)
		if err != nil && !os.IsExist(err) {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		tmp_path, err := os.MkdirTemp(path.Join(root_path, "tmp"), "unidock")
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		receptor_path := path.Join(tmp_path, "receptor.pdbqt")
		ligand_path := path.Join(tmp_path, "ligand.pdbqt")

		receptor_file, err := os.Create(receptor_path)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		defer receptor_file.Close()

		receptor_file.WriteString(strings.TrimSpace(receptor))

		receptor_file.Sync()

		ligand_file, err := os.Create(ligand_path)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		defer ligand_file.Close()

		ligand_file.WriteString(strings.TrimSpace(ligand))

		ligand_file.Sync()

		ligand_index_path := path.Join(tmp_path, "ligands.txt")
		ligand_index_file, err := os.Create(ligand_index_path)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		defer ligand_index_file.Close()

		ligand_index_file.WriteString(ligand_path)

		ligand_index_file.Sync()

		center_x := c.PostForm("center_x")
		center_y := c.PostForm("center_y")
		center_z := c.PostForm("center_z")

		size_x := c.PostForm("size_x")
		size_y := c.PostForm("size_y")
		size_z := c.PostForm("size_z")

		unidock_path, err := exec.LookPath("unidock")
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		cmd := exec.Command(unidock_path, "--receptor", receptor_path, "--ligand_index", ligand_index_path, "--center_x",
			center_x, "--center_y", center_y, "--center_z", center_z, "--size_x", size_x, "--size_y", size_y,
			"--size_z", size_z, "--scoring", "vina", "--search_mode", "balance", "--dir", tmp_path)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		fmt.Println(cmd.String())
		err = cmd.Run()
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error()+"\n"+stderr.String())
			return
		}

		ligand_out_path := path.Join(tmp_path, "ligand_out.pdbqt")
		ligand_out_file, err := os.Open(ligand_out_path)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		defer ligand_out_file.Close()

		scanner := bufio.NewScanner(ligand_out_file)
		foundModel := false
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "MODEL 1") {
				foundModel = true
				continue
			}

			if foundModel {
				regex := regexp.MustCompile(`^REMARK .+? RESULT:\s+(.+?)\s+(.+?)\s+(.+?)$`)
				match := regex.FindStringSubmatch(line)
				if len(match) > 0 {
					c.String(http.StatusOK, match[1])
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading file:", err)
		}

		c.String(http.StatusInternalServerError, "Error: No result found")
	})
	r.Run(":8080")
}
