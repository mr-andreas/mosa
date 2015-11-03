# Mosa

## parser

Converts a number of files into a raw AST. For instance, if we load the
following manifest, the AST will look exactly like it, with all variables and
everthing intact. Only syntax is checked. This means that the manifest may still
be invalid; undefined variables may be referenced, undeclared types may be
realized, cyclic dependendencies are allowed, etc.

```
node 'localhost' {
	class { 'webserver':
		wwwdir => "dev.local",
	}
}

class webserver($wwwdir,) {
	package { 'nginx': }
	
	file { "/etc/nginx/sites-available/server.conf":
		content => "
			server {
				root '$wwwdir'
				location / {}
			}
		", 
	}
}

define single package($name,) {
	exec { "apt-get install $name": }
}

define single file($name, $content,) {
	exec { "cat > $name":
		stdin => $content,
	}
}
```

## reducer

The reducer resolves all declarations in the manifest and returns them. The
declarations returned will all have concrete values. After our AST above has
been run through the reducer, the following will be returned:

```
exec { 'apt-get install nginx': }

package { 'nginx': }

exec { 'cat > /etc/nginx/sites-available/server.conf':
	stdin => '
		server {
			root 'dev.local'
			location / {}
		}
	',
}

file { '/etc/nginx/sites-available/server.conf':
	content => '
		server {
			root 'dev.local'
			location / {}
		}
	',
}
```

