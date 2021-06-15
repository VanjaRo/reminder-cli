package repositories

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reminders-cli/server/models"
)

type dbConfig struct {
	ID       int    `json:"id"`
	Checksum string `json:"checksum"`
}

// DB represents the application server database
type DB struct {
	dbPath    string
	dbCfgPath string
	cfg       dbConfig
	db        []byte
}

func NewDB(dbPath, dbCfgPath string) *DB {
	return &DB{
		dbPath:    dbPath,
		dbCfgPath: dbCfgPath,
	}
}

func (d *DB) Start() error {
	// Config path part
	bs, err := d.read(d.dbCfgPath)
	if err != nil {
		return models.WrapError("could not read db cfg file", err)
	}
	var cfg dbConfig
	if len(bs) == 0 {
		bs = []byte("{}")
	}
	err = json.Unmarshal(bs, &cfg)
	if err != nil {
		return models.WrapError("could not unmarshal db cfg file", err)
	}

	// DB path part
	bs, err = d.read(d.dbPath)
	if err != nil {
		return models.WrapError("could not read db contents", err)
	}
	d.db = bs
	if d.cfg.Checksum == "" {
		checksum, err := genChecksum(bytes.NewReader(bs))
		if err != nil {
			return err
		}
		cfg.Checksum = checksum
	}
	d.cfg = cfg

	return nil
}

func (d *DB) Read(bs []byte) (int, error) {
	n, err := bytes.NewReader(d.db).Read(bs)
	if err != nil {
		return 0, models.WrapError("could not read db file bytes", err)
	}

	return n, nil
}

func (d *DB) Write(bs []byte) (int, error) {
	bs = append(bs, '\n')
	checksum, err := genChecksum(bytes.NewReader(bs))
	if err != nil {
		return 0, err
	}
	if d.cfg.Checksum == checksum {
		return 0, nil
	}
	d.cfg.Checksum = checksum

	if err := d.writeDBCfg(); err != nil {
		return 0, err
	}

	n, err := d.write(d.dbPath, bs)
	if err != nil {
		return 0, err
	}
	d.db = bs

	return n, nil
}

func (d *DB) Size() int {
	if len(d.db) == 0 {
		d.db = []byte("{}")
	}
	return len(d.db)
}

func (d *DB) GenerateID() int {
	d.cfg.ID++
	return d.cfg.ID
}

func (d DB) Stop() error {
	log.Println("shutting down db")
	_, errDB := os.Open(d.dbPath)
	if errors.Is(errDB, os.ErrNotExist) {
		_, err := d.write(d.dbPath, d.db)
		if err != nil {
			return err
		}
	}
	_, errCnf := os.Open(d.dbCfgPath)
	if errors.Is(errCnf, os.ErrNotExist) {
		err := d.writeDBCfg()
		if err != nil {
			return err
		}
	}
	log.Println("db was successfully shut down")
	return nil
}

func (d *DB) read(path string) ([]byte, error) {
	dbFile, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
	if errors.Is(err, os.ErrNotExist) {
		dbFile, err = os.Create(path)
	}
	if err != nil {
		return nil, models.WrapError("could not open or create db file", err)
	}
	return ioutil.ReadAll(dbFile)

}

func (d *DB) write(path string, bs []byte) (int, error) {
	dbFile, err := os.Create(path)
	if err != nil {
		return 0, models.WrapError("could not create file", err)
	}
	defer d.close(dbFile)

	n, err := dbFile.Write(bs)
	if err == nil {
		log.Printf("successfully wrote %d byte(s) to %s file", n, dbFile.Name())
	}
	return n, err
}

func (d *DB) close(f *os.File) {
	if err := f.Close(); err != nil {
		log.Printf("could not close file %s: %v", f.Name(), err)
	}
}

func (d *DB) writeDBCfg() error {
	bs, err := json.Marshal(d.cfg)
	if err != nil {
		return models.WrapError("could not marshal db cfg", err)
	}
	bs = append(bs, '\n')
	_, err = d.write(d.dbCfgPath, bs)
	if err != nil {
		return models.WrapError("could not write to db cfg file", err)
	}
	return nil
}

func genChecksum(r io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, r); err != nil {
		return "", models.WrapError("could not copy db content", err)
	}
	sum := hash.Sum(nil)
	return fmt.Sprint("%x", sum), nil
}
