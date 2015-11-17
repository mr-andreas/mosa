define single file($name, $ensure, $content = "",) {
	exec { "cat > $name":
		stdin => $content,
	}
}
