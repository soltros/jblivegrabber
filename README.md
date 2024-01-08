# jblivegrabber
A quick and dirty tool to quickly grab the JB livestreams and convert them to MP3s in an XML feed.

To use: Install yt-dlp, and edit jb-livestream-grabber.go and change ``enclosureURL := "http://YOUR_WEB_SERVER/jblivegrabber/podcasts/" + url.PathEscape(f.Name()) // Replace with actual URL`` to your webserver's hostname.

Next, run:

``go mod init jb-livestream-grabber.go``

``go build jb-livestream-grabber.go``

``./jb-livestream-grabber``
