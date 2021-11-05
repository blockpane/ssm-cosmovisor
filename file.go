package scv

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"strings"
	"syscall"
	"time"
)

const keyfile = `/config/priv_validator_key.json` // relative to DAEMON_HOME env

var (
	dHome = strings.TrimRight(os.Getenv("DAEMON_HOME"), `/`)
	pipeTimeout = 15*time.Minute // how long to wait for a reader on our pipe. This is long just in case of a backup
)

// WritePipeOnce creates a named unix pipe/aka FIFO and provides the Private Key once then exits,
// once it has exited, WriteStrippedKey should be called.
func WritePipeOnce(pk *PrivValKey) error {
	j, err := json.MarshalIndent(pk, "", "  ")
	if err != nil {
		return err
	}

	_ = os.Remove(dHome+keyfile)
	err = syscall.Mkfifo(dHome+keyfile, 0600)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(dHome+keyfile, os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	err = f.SetWriteDeadline(time.Now().Add(pipeTimeout))
	if err != nil {
		return err
	}
	log.Println("writing key to named pipe")
	_, err = f.Write(j)
	if err != nil {
		return err
	}
	//_, err = f.Write([]byte{0})
	log.Println("key was written to named pipe")
	err = os.Remove(dHome+keyfile)
	return err
}

// WriteStrippedKey will write a regular file containing only the public key
func WriteStrippedKey(pk PrivValKey) error {
	pk.PrivKey.Value = "" // removes the private key
	stripped, err := json.MarshalIndent(pk, "", "  ")
	if err != nil {
		return err
	}
	key, _ := os.Open(dHome+keyfile)
	// cleanup if our pipe still exists
	if key != nil {
		fi, _ := key.Stat()
		_ = key.Close()
		if fi.Mode() == os.ModeNamedPipe {
			err = os.Remove(dHome + keyfile)
			if err != nil {
				return err
			}
		}
	}
	sk, err := os.OpenFile(dHome+keyfile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, err = sk.Write(stripped)
	if err != nil {
		_ = sk.Close()
		return err
	}
	log.Println("writing stripped key file")
	return sk.Close()
}

// BackupOrig saves the original key to a backup file and unlinks the original file.
func BackupOrig() error {
	log.Println("backing up original key")
	if dHome == "" {
		log.Fatal("env var DAEMON_HOME must be set, exiting")
	}
	orig, err := os.Open(dHome+keyfile)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(orig)
	_ = orig.Close()
	if err != nil {
		return err
	} else if data == nil {
		return errors.New(keyfile+" was empty")
	}

	err = os.Remove(dHome+keyfile)
	if err != nil {
		return err
	}

	// if we run into trouble backing up the key, restore the original:
	undoRemove := func() {
		log.Println("backup failed: attempting to restore original key")
		restore, e := os.OpenFile(dHome+keyfile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if e != nil {
			log.Println(e)
			return
		}
		_, e = restore.Write(data)
		if e != nil {
			log.Println(e)
			return
		}
		e = restore.Close()
		if e != nil {
			log.Println(e)
		}
	}

	backup, err := os.OpenFile(dHome+keyfile+`.orig`, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		undoRemove()
		return err
	}
	defer backup.Close()

	_, err = backup.Write(data)
	if err != nil {
		undoRemove()
		return err
	}

	log.Println("original key saved")
	return nil
}

// RestoreOrig copies the originally present priv_validator_key.json back in place.
// this allows having a "safe" key in place, and only acting as a validator when
// the 'USE_SSM_KEY' env var is true.
func RestoreOrig() error {
	log.Println("attempting to restore original key file")
	backup, err := os.OpenFile(dHome+keyfile+`.orig`, os.O_RDONLY, 0600)
	if err != nil {
		return err
	}
	original, err := io.ReadAll(backup)
	_ = backup.Close()
	if err != nil {
		return err
	} else if original == nil || len(original) == 0 {
		return errors.New("backup key file was empty, not restoring")
	}

	restored, err := os.OpenFile(dHome+keyfile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer restored.Close()
	_, err = restored.Write(original)
	if err == nil {
		log.Println("restored original key file")
	}
	return err
}
