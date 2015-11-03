define single user($name, $password,) {
	exec { "useradd -p '$password' $name": }
}
