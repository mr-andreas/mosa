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

## resolver

The resolver resolves all declarations in the manifest and returns them. The
declarations returned will all have concrete values. This gives us a definition
of how the final state of our target system should look like. 

After our AST above has been run through the resolver, the following will be
returned:

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

## facter

Now that we have a clear image of what the final state of the target system
should be, it's time to run the facter. The facter helps us find out which
declarations are already fullfilled in the target system. With this knowledge,
we can remove those declarations from the manifest, as we need no action to
reach them.

If we for example's sake suppose that `/etc/nginx/sites-available/server.conf`
already exists on the target system, and has the content specified, our
declaration list would look like the following after the facter has been
invoked:

```
exec { 'apt-get install nginx': }

package { 'nginx': }
```

## stepconverter

After the facter has been invoked, we only have a subset of our original
declarations that we need to actually execute. The stepconverter converts these
into a number of concrete steps with dependencies between them. As `exec` is the
only builtin type in mosa, all defines result in a number of `exec`
declarations. It's these declarations that are converted into steps, while
everything else is thrown away:

```
exec { 'apt-get install nginx': }
```

## planner

The planner resolves the dependencies between the steps that we need to execute,
and groups them into a number of stages. 

