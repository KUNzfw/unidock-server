package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.POST("/unidock", func(c *gin.Context) {
		receptor := c.PostForm("receptor")
		ligand := c.PostForm("ligand")

		err := os.Mkdir("tmp", 0755)
		if err != nil && !os.IsExist(err) {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		tmp_path, err := os.MkdirTemp("tmp", "unidock")
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

		receptor_file.WriteString(receptor)

		ligand_file, err := os.Create(ligand_path)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		defer ligand_file.Close()

		ligand_file.WriteString(ligand)

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

		cmd := exec.Command(unidock_path, "--receptor", receptor_path, "--ligand", ligand_path, "--center_x",
			center_x, "--center_y", center_y, "--center_z", center_z, "--size_x", size_x, "--size_y", size_y,
			"--size_z", size_z, "--scoring", "vina", "--search_mode", "balance")

		fmt.Println(cmd.String())
		err = cmd.Run()
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		c.String(http.StatusOK, "Success!")
	})
	r.Run(":8080")
}
