CKEditor-oEmbed-Plugin
======================

oEmbed Plugin for CKEditor

This Plugin allows to insert embedded content (such as photos, videos, audio, and other rich media) via the OEmbed API. You only have to provide the url to the site (It works also when the url is shortened) you want to embed and the plugin does the rest.

This Plugin uses the jquery-oembed-all Plugin  located at https://github.com/starfishmod/jquery-oembed-all.

####Demo

http://w8tcha.github.com/CKEditor-oEmbed-Plugin/

####Currently Supported Sites...

###Video
* Youtube - oembed - YQL
* Blip - oEmbed
* Hulu - oEmbed
* Vimeo - oEmbed
* National film board of Canada - oEmbed
* Qik - oEmbed
* Dotsub - oEmbed
* Clikthrough - oEmbed
* Kino Map - oEmbed
* Funny Or Die - Embedded
* College Humour - Embedded
* Metacafe - Embedded
* embedr - Embedded
* 5min - oEmbed is XML only - using YQL to translate it
* ustream.tv - oEmbed is not JSONP enabled - using YQL to translate it
* viddler - OGP
* twitvid - Embedded
* bambuser - Embedded
* xtranormal - Embedded
* Gametrailers - Embedded
* Vzarr - Embedded
* VHX - oembed
* bambuser - oembed
* dailymotion.com - oembed
* animoto - oembed
* justin.tv - YQL JSON
* livestream - OGP
* scivee - embedded
* veoh - embedded
* minoto-video - oembed using YQL
* TrailerAddict - OGP
* vodpod - oembed YQL - broken as the oembed has absolute positioning which breaks the display
* fora.tv -OGP YQL
* TED - OGP YQL
* Aniboom - embedded
* Comedy Central - OGP
* snotr - embedded
* zapiks - OGP
* youku - embedded
* wistia - Oembed

###Audio
* Soundcloud - oEmbed
* HuffDuffer - oEmbed
* BandCamp - YQL and Embedded
* podomatic - OGP
* rdio.com - oEmbed
* hark.com - OGP
* chirb.it - YQL and oembed
* official.fm - YQL and oembed
* mixcloud - YQL and oembed
* shoudio - oembed
* audioboo.fm - OGP 
* Spotify - OGP YQL

###Photo
* flickr - oEmbed
* photobucket - oEmbed
* instagram - oEmbed
* yfrog - oEmbed
* 23HQ - oEmbed
* Smugmug - oEmbed
* twitpic - OGP YQL
* 500px.com - OGP
* visual.ly - YQL Lookup
* img.ly - Thumbnail view
* imgur.com - Thumbnail view
* twitgoo.com - Thumbnail view
* gravatar - Thumbnail view when using mailto
* pintrest - YQL - Embedded view of a sort.
* circuitlab - image view
* skitch - YQL oembed
* graphic.ly  - OGP
* dribble - jsonp lookup
* Lockerz - YQL lookup
* AsciiArtFarts - YQL Lookup
* lego cusoo - OGP over YQL
* plannary - OGP over YQL
* propic - OGP
* avairy.com - OGP
* lomography - ogp
* weheartit - ogp 
* glogster - ogp
* chart.ly - embedded
* twitrpix - OGP
* chictopia - OGP

###Rich
* Meetup - oEmbed
* gigapans - Embedded
* Slideshare - oEmbed
* ebay - Embedded
* scribd - Embedded
* screenr - Embedded
* tumblr- JSONP lookup
* imdb - JSONP lookup via imdbapi.com
* wikipedia- JSONP lookup
* github- JSONP lookup (CSS) 
* eventful - OGP
* myspace - OGP
* live Journal - JSONP Lookup (CSS)
* wordpress - oEmbed (wordpress.com, wp.me, blogs.cnn.com, techcrunch.com). I can add other wordpress sites as well.
* circuitbee -Embedded
* stack overflow - JSONP Lookup (CSS)
* Facebook - JSONP Lookup (CSS)
* Pastebin - Embedded
* Pastie - YQL lookup
* kickstarter - Embedded
* issuu - OGP
* reelapp.com - Embedded
* Etsy - OGP over YQL
* Amazon - Embedded - Requires Affiliate code
* linkedin - Embedded IFRAME - found a link that works :)
* Lanyrd - YQL (CSS)
* twitter - Oembed - status only - but that is ok I think
* github gist - oembed
* speakerdeck - yql oembed
* dipity - yql oembed
* dailymile - oembed
* deviantart - oembed
* Roomshare Japan - oembed
* mobypictures - oembed
* prezi - embedded
* popplet - embedded
* authorstream - OGP
* googlecalendar - Iframe
* cacoo - oembed
* pearltrees - embedded
* urtak - oembed - is broken in iframe return atm -seems to be an embed.ly issue??
* jotform - embedded
* Urban Dictionary - YQL lookup
* Ars Technica - YQL Lookup
* Eventbrite - OGP YQL
* last.fm OGP YQL
* Rotten Tomatoes - OGP YQL
* iFixit - OGP
* qwiki - OGP
* brighttalk - Meta info
* tinychat - OGP
* tourwrist - embedded
* bnter - OGP
* bigthink - OGP
* wirewax - OGP
* whosay - OGP
* timetoast - embedded
* tripline - OGP
* jsfiddle - embedded


####License

Licensed under the terms of the MIT License.

####Dependencies 
This Plugin requires the following plugins to work: Widget, Dialog.

####Installation

 1. Before you can use the plugin you also need to download & install the widget plugin, if you have it not installed. http://ckeditor.com/addon/widget
 2. Extract the contents of the file into the "plugins" folder of CKEditor.
 3. In the CKEditor configuration file (config.js) add the following code:

````js
config.extraPlugins = 'oembed,widget';
````

2a. Additionally you can also set the default values vor the Max. Width/Height Values

````js
config.oembed_maxWidth = '560';
config.oembed_maxHeight = '315';
````

and also you can define an css class for the embeded content wrapper (div), by default there is no Class defined

````js
config.oembed_WrapperClass = 'embededContent';
````


3. and also include the plugin in the toolbar 

````js
toolbar :[ ... ['oembed']...]
````
