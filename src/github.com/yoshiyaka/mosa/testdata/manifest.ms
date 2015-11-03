node 'localhost' {
	class { 'WebServer': }
}

class WebServer {
	exec { 'apt-get update': }
	
	package { 'apache2':
	  depends => exec['apt-get update'],
	}
	
	package { 'php5':
	  depends => exec['apt-get update'],
	}
	
	package { 'proftpd':
	  depends => exec['apt-get update'],
	}
	
	user { 'ftpuser':
	  password => 'test',
	}
	
	file { '/etc/apache2/sites-available/test-site.conf':
	  ensure => 'present',
	  depends => package['apache2'],
	}
	
	exec { 'a2ensite test-site':
	  depends => file['/etc/apache2/sites-available/test-site.conf'],
	}
}
