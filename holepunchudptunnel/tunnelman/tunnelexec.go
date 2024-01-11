package tunnelman

import (
	"embed"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
	"path/filepath"
	"go.uber.org/zap"
)

// Embed UDPTunnel executable inside of our HolePunch-UDPTunnel executable

//go:embed udptunnel
var embeddedExecutable embed.FS

// Execute embedded executable with sudo privillidges
func executeEmbeddedBinaryAsSudo(l *zap.Logger, embedded embed.FS, binaryPath, configFile, password string, args ...string) {
	// Extraction of embedded binary and execution logic
	tmpBinary, err := ioutil.TempFile("", "embedded-executable-*")
	if err != nil {
		l.Fatal("Error creating temp file:"+err.Error())
		return
	}
	defer os.Remove(tmpBinary.Name()) // Cleanup: Remove the temporary file on program exit

	data, err := embedded.ReadFile(binaryPath)
	if err != nil {
		l.Fatal("Error reading embedded file:"+err.Error())
		return
	}

	if _, err := tmpBinary.Write(data); err != nil {
		l.Fatal("Error writing temp file:"+err.Error())
		return
	}

	// Allow execution permission to the temp file
	if err := tmpBinary.Chmod(0755); err != nil {
		l.Fatal("Error setting file permission:"+err.Error())
		return
	}

	// Close tmpBinary so sudo can read it
	tmpBinary.Close()
	tempExecPath, err := filepath.Abs(tmpBinary.Name())
	if err != nil {
		l.Fatal("Error getting path to tmpBinary:"+err.Error())
	}
	l.Info("Extracted udptunnel executable to "+tempExecPath)

	// Copy config file to the extracted executable dir
	tempExecDirPath, err := filepath.Abs(filepath.Dir(tmpBinary.Name()))
	if err != nil {
		l.Fatal("Error getting path to tmpBinary folder:"+err.Error())
	}
	copyFileToTempDir(l, configFile, tempExecDirPath)
	l.Info("Copied config file to "+tempExecDirPath)

	// Prepare sudo command to execute the extracted binary with arguments
	cmdArgs := append([]string{"-S", tmpBinary.Name()}, args...)
	sudoCmd := exec.Command("sudo", cmdArgs...)

	// Create pipes for stdin, stdout, and stderr
	stdin, err := sudoCmd.StdinPipe()
	if err != nil {
		l.Fatal("Error creating stdin pipe:"+err.Error())
		return
	}
	defer stdin.Close()

	stdout, err := sudoCmd.StdoutPipe()
	if err != nil {
		l.Fatal("Error creating stdout pipe:"+err.Error())
		return
	}

	stderr, err := sudoCmd.StderrPipe()
	if err != nil {
		l.Fatal("Error creating stderr pipe:"+err.Error())
		return
	}

	// Start the sudo command
	if err := sudoCmd.Start(); err != nil {
		l.Fatal("Error starting sudo command:"+err.Error())
		return
	}

	// Write sudo password to stdin
	_, err = io.WriteString(stdin, password+"\n")
	if err != nil {
		l.Fatal("Error writing to stdin:"+err.Error())
		return
	}

	// Print stdout and stderr in real-time
	go printOutput(l, stdout)
	go printOutput(l, stderr)

	// Wait for the sudo command to finish
	if err := sudoCmd.Wait(); err != nil {
		l.Fatal("Error running sudo command:"+err.Error())
		return
	}
	//! when exiting with ctrl+c the signal is not passed to the running executable

	// Remove the copied config file after embedded binary execution
	removeCopiedFile(l, filepath.Join(tempExecDirPath, filepath.Base(configFile)))
}

func printOutput(l *zap.Logger, pipe io.Reader) {
	buf := make([]byte, 4096)
	for {
		n, err := pipe.Read(buf)
		if err != nil && err != io.EOF {
			l.Error("Error reading from pipe:"+err.Error())
			return
		}
		if n > 0 {
			l.Info(string(buf[:n]))
		}
		// if err == io.EOF {
		// 	return
		// }
	}
}

func copyFileToTempDir(l *zap.Logger, fileToCopy, destDir string) {
	srcFile, err := os.Open(fileToCopy)
	if err != nil {
		l.Fatal("Error opening source file:"+err.Error())
		return
	}
	defer srcFile.Close()

	destFile, err := os.Create(filepath.Join(destDir, filepath.Base(fileToCopy)))
	if err != nil {
		l.Fatal("Error creating destination file:"+err.Error())
		return
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		l.Fatal("Error copying file:"+err.Error())
		return
	}

	l.Info("File copied successfully.")
}

func removeCopiedFile(l *zap.Logger, filePath string) {
	err := os.Remove(filePath)
	if err != nil {
		l.Fatal("Error removing copied file:"+err.Error())
		return
	}
	l.Info("Copied file removed successfully.")
}

func forwardSignalToChild(l *zap.Logger, sig os.Signal) {
	// Get PID of child executable
	pid := os.Getpid()

	// Create a process group with the same process group ID as the parent
	err := syscall.Setpgid(pid, pid)
	if err != nil {
		l.Fatal("Error setting process group ID:"+err.Error())
		return
	}

	// Forward the signal to the child process group
	err = syscall.Kill(-pid, sig.(syscall.Signal))
	if err != nil {
		l.Error("Error forwarding signal to child process:"+err.Error())
	}
}