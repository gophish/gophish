SlideShow
=========
SlideShow Plugin for CKEditor

### New Feature
**December 12, 2015 : NEW Feature which allows to add "clickable" link on images, 
look at the [documentation](http://devlabnet.eu/softdev/slideshow/index.php) for more information.**

A cool plugin which allow to create and manage SlideShow in CKEditor.
You can easily Add, Remove images to create the Slide Show.

Specification
-------------
The plugin has been designed to work with the "Ad-Gallery" javascript slidshow program available at the
following location : http://adgallery.codeplex.com, and with "FancyBox" java program available at the
following location : http://fancybox.net/.

For each slide show created with this plugin, you can adjust most of the available controls
available in ad-gallery :

    Slide Effect.
    Animation Speed.
    Animation Delay.
    Auto Start
    Show / Hide Thumbnails.
    Sho / Hide "Start - Stop" Buttons
    Open Image on Click (with a FancyBox pop-up).
    ...

Internationalization
-------------------------
Currently plugin supports 2 languages.

* en
* fr
* ru Russian
* el Greek, Modern (1453-)
* sr Serbian
* sr-latin
* pt Portuguese
* pt-br Brazilian Portuguese
*Translations are welcomed.*

Usage
-------------------------
1. Download source files and place them on to be created "slideshow" folder under the CKeditor's plugin base.
2. Define plugin in CKEDITOR config object.
        CKEDITOR.config.extraPlugins = 'slideshow';
3. Set your CKEDITOR language if you did not set it yet.
        CKEDITOR.config.language = 'en';
4. You're Done !! Just enjoy.

The needed files for "ad-gallery" and "fancybox" are located under the 3rdParty directory, in the plugin package.
They are just copy of the files from respective web sites (ad-gallery and fancybox), just a fews modifs have been made
in the as-gallery css file, for info, and for curious people, the diffs compared to the original are in the patch
file under this ad-gallery directory.
Normally, nothing special has too be done with these files. If you like to change their location, just edit the "slidesho.js"
 and update the variables on top of this file.

Requirements
-------------------------
To correctly work, you need to have access to CKEditor, KCFinder (or any stuff to allow to upkoad images
on the server), ad-gallery javascript / css and fancybox javascript and css.

Demo
-------------------------
[View the live demo]( http://devlabnet.eu/softdev/slideshow/demo.php ).


Cheers
--------------------
Thanks to [CKeditor] [1] and [ad-gallery] [2] and [fancybox] [3] people for their good work.

  [1]: http://ckeditor.com              "CKeditor"
  [2]: http://adgallery.codeplex.com    "ad-gallery"
  [3]: http://fancybox.net/             "fancybox"
