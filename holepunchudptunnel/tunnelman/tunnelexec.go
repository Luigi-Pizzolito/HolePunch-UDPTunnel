package tunnelman

// import (
// 	"embed"
// 	"fmt"
// 	"io"
// 	"io/ioutil"
// 	"os"
// 	"os/exec"
// 	"path/filepath"
// )

// //go:embed embedded/executable
// var embeddedExecutable embed.FS

// // func main() {
// // 	password := "your_sudo_password" // Replace with your sudo password
// // 	args := []string{"arg1", "arg2"}  // Replace with the arguments to be passed

// // 	// Path to the file to be copied (relative to the Go executable)
// // 	fileToCopy := "path/to/file.ext" // Replace with the relative path of your file

// // 	// Get the absolute path to the Go executable
// // 	exePath, err := os.Executable()
// // 	if err != nil {
// // 		fmt.Println("Error getting executable path:", err)
// // 		return
// // 	}
// // 	exeDir := filepath.Dir(exePath)

// // 	// Copy the fileToCopy to the same directory as the temporary binary
// // 	copyFileToTempDir(fileToCopy, exeDir)

// // 	// Extract and execute the embedded binary with sudo
// // 	executeEmbeddedBinaryAsSudo(embeddedExecutable, "embedded/executable", password, args...)

// // 	// Remove the copied file after embedded binary execution
// // 	removeCopiedFile(filepath.Join(exeDir, filepath.Base(fileToCopy)))
// // }

// func executeEmbeddedBinaryAsSudo(embedded embed.FS, binaryPath, password string, args ...string) {
// 	// Extraction of embedded binary and execution logic
// 	tmpBinary, err := ioutil.TempFile("", "embedded-executable-*")
// 	if err != nil {
// 		fmt.Println("Error creating temp file:", err)
// 		return
// 	}
// 	defer os.Remove(tmpBinary.Name()) // Cleanup: Remove the temporary file on program exit

// 	data, err := embedded.ReadFile(binaryPath)
// 	if err != nil {
// 		fmt.Println("Error reading embedded file:", err)
// 		return
// 	}

// 	if _, err := tmpBinary.Write(data); err != nil {
// 		fmt.Println("Error writing temp file:", err)
// 		return
// 	}

// 	// Allow execution permission to the temp file
// 	if err := tmpBinary.Chmod(0755); err != nil {
// 		fmt.Println("Error setting file permission:", err)
// 		return
// 	}

// 	// Prepare sudo command to execute the extracted binary with arguments
// 	cmdArgs := append([]string{"-S", tmpBinary.Name()}, args...)
// 	sudoCmd := exec.Command("sudo", cmdArgs...)

// 	// Create pipes for stdin, stdout, and stderr
// 	stdin, err := sudoCmd.StdinPipe()
// 	if err != nil {
// 		fmt.Println("Error creating stdin pipe:", err)
// 		return
// 	}
// 	defer stdin.Close()

// 	stdout, err := sudoCmd.StdoutPipe()
// 	if err != nil {
// 		fmt.Println("Error creating stdout pipe:", err)
// 		return
// 	}

// 	stderr, err := sudoCmd.StderrPipe()
// 	if err != nil {
// 		fmt.Println("Error creating stderr pipe:", err)
// 		return
// 	}

// 	// Start the sudo command
// 	if err := sudoCmd.Start(); err != nil {
// 		fmt.Println("Error starting sudo command:", err)
// 		return
// 	}

// 	// Write sudo password to stdin
// 	_, err = io.WriteString(stdin, password+"\n")
// 	if err != nil {
// 		fmt.Println("Error writing to stdin:", err)
// 		return
// 	}

// 	// Print stdout and stderr in real-time
// 	go printOutput(stdout)
// 	go printOutput(stderr)

// 	// Wait for the sudo command to finish
// 	if err := sudoCmd.Wait(); err != nil {
// 		fmt.Println("Error running sudo command:", err)
// 		return
// 	}
// }

// func printOutput(pipe io.Reader) {
// 	buf := make([]byte, 4096)
// 	for {
// 		n, err := pipe.Read(buf)
// 		if err != nil && err != io.EOF {
// 			fmt.Println("Error reading from pipe:", err)
// 			return
// 		}
// 		if n > 0 {
// 			fmt.Print(string(buf[:n]))
// 		}
// 		if err == io.EOF {
// 			return
// 		}
// 	}
// }

// func copyFileToTempDir(fileToCopy, destDir string) {
// 	srcFile, err := os.Open(fileToCopy)
// 	if err != nil {
// 		fmt.Println("Error opening source file:", err)
// 		return
// 	}
// 	defer srcFile.Close()

// 	destFile, err := os.Create(filepath.Join(destDir, filepath.Base(fileToCopy)))
// 	if err != nil {
// 		fmt.Println("Error creating destination file:", err)
// 		return
// 	}
// 	defer destFile.Close()

// 	_, err = io.Copy(destFile, srcFile)
// 	if err != nil {
// 		fmt.Println("Error copying file:", err)
// 		return
// 	}

// 	fmt.Println("File copied successfully.")
// }

// func removeCopiedFile(filePath string) {
// 	err := os.Remove(filePath)
// 	if err != nil {
// 		fmt.Println("Error removing copied file:", err)
// 		return
// 	}
// 	fmt.Println("Copied file removed successfully.")
// }
