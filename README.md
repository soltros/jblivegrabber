# jblivegrabber
A quick and dirty tool to quickly grab the JB livestreams and convert them to MP3s in an XML feed.

To use: Install yt-dlp, and edit jb-livestream-grabber.go and change ``enclosureURL := "http://YOUR_WEB_SERVER/jblivegrabber/podcasts/" + url.PathEscape(f.Name()) // Replace with actual URL`` to your webserver's hostname.

Next, run:

``go mod init jb-livestream-grabber.go``

``go mod init xml_feed_cleaner.go``

``go build jb-livestream-grabber.go``

``go build xm_feed_cleaner.go``

``./jb-livestream-grabber``

To sanatize the XML feed post generation, run:

``./xm_feed_cleaner``
