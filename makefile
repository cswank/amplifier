install:
	sudo mount $(dev) /mnt/pico/
	tinygo build -o main.uf2 -target=pico main.go
	sudo cp main.uf2 /mnt/pico/
	sudo umount /mnt/pico


