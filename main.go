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
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Vec3 struct {
	x float64
	y float64
	z float64
}

func main() {
	r := gin.Default()

	r.POST("/unidock", func(c *gin.Context) {
		receptor_formfile, err := c.FormFile("receptor")
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		ligand_formfile, err := c.FormFile("ligand")
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

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

		receptor_path := path.Join(tmp_path, receptor_formfile.Filename)
		ligand_path := path.Join(tmp_path, ligand_formfile.Filename)

		c.SaveUploadedFile(receptor_formfile, receptor_path)
		c.SaveUploadedFile(ligand_formfile, ligand_path)

		center, size, err := searchPocket(receptor_path)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		ligand_index_path := path.Join(tmp_path, "ligands.txt")
		ligand_index_file, err := os.Create(ligand_index_path)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		defer ligand_index_file.Close()

		ligand_index_file.WriteString(ligand_path)

		ligand_index_file.Sync()

		unidock_path, err := exec.LookPath("unidock")
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		cmd := exec.Command(unidock_path, "--receptor", receptor_path, "--ligand_index", ligand_index_path, "--center_x",
			fmt.Sprint(center.x), "--center_y", fmt.Sprint(center.y), "--center_z", fmt.Sprint(center.z), "--size_x",
			fmt.Sprint(size.x), "--size_y", fmt.Sprint(size.y), "--size_z", fmt.Sprint(size.z), "--scoring", "vina",
			"--search_mode", "balance", "--dir", tmp_path)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		fmt.Printf("Running command: %s\n", cmd.String())

		err = cmd.Run()
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error()+"\n"+stderr.String())
			return
		}

		ligand_out_path := path.Join(tmp_path, strings.TrimSuffix(ligand_formfile.Filename, path.Ext(ligand_formfile.Filename))+"_out.pdbqt")
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

func searchPocket(receptor_path string) (center Vec3, size Vec3, err error) {
	//dir: path to the directory containing the receptor file
	//receptor_name: name of the receptor file

	dir := path.Dir(receptor_path)
	receptor_name := strings.TrimSuffix(path.Base(receptor_path), path.Ext(receptor_path))
	fpocket_path, err := exec.LookPath("fpocket")

	if err != nil {
		return
	}

	cmd := exec.Command(fpocket_path, "-f", receptor_path)
	err = cmd.Run()
	if err != nil {
		return
	}

	pocket1_path := path.Join(dir, receptor_name+"_out", "pockets", "pocket1_atm.pdb")
	pocket1_file, err := os.Open(pocket1_path)
	if err != nil {
		return
	}
	defer pocket1_file.Close()

	var pos []Vec3

	scanner := bufio.NewScanner(pocket1_file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ATOM") {
			var x, y, z float64
			x, err = strconv.ParseFloat(strings.TrimSpace(line[30:38]), 64)
			if err != nil {
				return
			}
			y, err = strconv.ParseFloat(strings.TrimSpace(line[38:46]), 64)
			if err != nil {
				return
			}
			z, err = strconv.ParseFloat(strings.TrimSpace(line[46:54]), 64)
			if err != nil {
				return
			}
			pos = append(pos, Vec3{x, y, z})
		}
	}

	if err = scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	for i := 0; i < len(pos); i++ {
		center.x += pos[i].x
		center.y += pos[i].y
		center.z += pos[i].z
	}

	center.x /= float64(len(pos))
	center.y /= float64(len(pos))
	center.z /= float64(len(pos))

	// max
	max := pos[0]
	for i := 1; i < len(pos); i++ {
		if pos[i].x > max.x {
			max.x = pos[i].x
		}
		if pos[i].y > max.y {
			max.y = pos[i].y
		}
		if pos[i].z > max.z {
			max.z = pos[i].z
		}
	}

	// min
	min := pos[0]
	for i := 1; i < len(pos); i++ {
		if pos[i].x < min.x {
			min.x = pos[i].x
		}
		if pos[i].y < min.y {
			min.y = pos[i].y
		}
		if pos[i].z < min.z {
			min.z = pos[i].z
		}
	}

	// size
	size.x = max.x - min.x
	size.y = max.y - min.y
	size.z = max.z - min.z

	return
}
