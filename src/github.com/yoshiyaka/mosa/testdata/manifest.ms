node 'localhost' {
	class { 'WebServer': }
}

class WebServer {
	shell { 'apt-get update': }
	
	package { 'apache2':
	  depends => shell['apt-get update'],
	}
	
	package { 'php5':
	  depends => shell['apt-get update'],
	}
	
	package { 'proftpd':
	  depends => shell['apt-get update'],
	}
	
	user { 'ftpuser':
	  password => 'test',
	}
	
	file { '/etc/apache2/sites-available/test-site.conf':
	  ensure => 'present',
	  depends => package['apache2'],
	}
	
	shell { 'a2ensite test-site':
	  depends => file['/etc/apache2/sites-available/test-site.conf'],
	}
}
