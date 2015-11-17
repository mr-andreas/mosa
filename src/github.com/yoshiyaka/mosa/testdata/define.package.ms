define single package($name,) {
	exec { "apt-get -f install -y $name": }
}
