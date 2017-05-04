Let's assume the following:

- Your nginx is running under the document root of `/var/www/`
- You install the `gobot` binary in that directory
- `/lib` and `/templates` are placed under `/var/www/lib` and `/var/www/templates`, respectively
- `gobot` is set to run on localhost at the default port 3000
- Besides serving the files statically via nginx (Go can do it too, but nginx is more efficient), you wish to use nginx's internal super-fast and efficient caching system for the replies coming from Go; a possible place to store those cache files is under `/var/cache/nginx/`

Then the configuration below is what you need to add for that server (or virtual host), replacing, of course, the `root` directives with `/var/www/lib` and `/var/www/templates`.

You can, of course, put those directories elsewhere (e.g. a read-only directory away from the main server, whatever) but then remember to update `PathToStaticFiles` in `config.toml`. Currently those need to be in the same server, but, who knows, I might add the extra logic to host those static files elsewhere. For my purposes, nginx is already super-fast...

```
[... lots of your existing config options ...]

http {
	# proxy configuration, adjust as needed
	proxy_temp_path		/var/cache/nginx/temp;
	proxy_cache_path	/var/cache/nginx/cached  levels=1:2    keys_zone=STATIC:15m inactive=24h  max_size=1g;
	proxy_cache_valid	200 302 1d;
	proxy_cache_valid	301 1h;
	proxy_cache_valid	404 3m;    
	proxy_cache_use_stale error timeout invalid_header updating http_500 http_502 http_503 http_504; 
	proxy_cache_key		"$scheme$host$request_uri $cookie_user";

[... even more lots of your existing config options ...]
}

server {

[... blah blah blah blah ...]

	location ~* /go/lib {
		rewrite ^/go/lib/(.*)$ /$1 break;
		root [my directory where I have my lib files (CSS, JS, etc.)];
		try_files $uri =404;
	}
	
	location ~* /go/templates {
		rewrite ^/go/templates/(.*)$ /$1 break;
		root [my directory where I have my templates];
		try_files $uri =404;
	}
	
	location /go {
		add_header X-Proxy-Cache $upstream_cache_status;
		proxy_set_header X-Real-IP $remote_addr;
		proxy_set_header X-Forwarded-For $remote_addr;
		proxy_set_header Host $host;
		proxy_cache STATIC; # if you wish to use nginx's super-efficient caching system
		proxy_pass_header Set-Cookie;
		proxy_pass http://127.0.0.1:3000;
	}
	
	location ~* ^.+.(jpg|jpeg|gif|png|ico|css|zip|tgz|gz|rar|bz2|doc|xls|exe|pdf|ppt|txt|tar|tpl|mid|midi|wav|bmp|rtf|js|swf|flv|mp3)$ {
		access_log off;
		add_header Cache-Control "public";
		proxy_cache_valid 200 1d;
		expires 3d;
	}
	
[... blah, blah, blah ...]
	
}
```

<< Back to [main readme](README.md)