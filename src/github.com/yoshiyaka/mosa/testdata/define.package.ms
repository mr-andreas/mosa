define single package($name,) {
	exec { "apt-get -f install $name": }
}
