BootstrapCK Skin
====================

The BootstrapCK-Skin is a skin for [CKEditor4](http://ckeditor.com/) based on [Twitter Bootstrap3](http://getbootstrap.com/) styles.

[Sass](http://sass-lang.com/) is used to rewrite the editor's styles and [Grunt](http://gruntjs.com/) to be able to watch, convert and minify the sass into css files. These files aren't really needed for the simple use of the skin, but handy if you want to make some adjustments to it.

For more information about skins, please check the [CKEditor Skin SDK](http://docs.cksource.com/CKEditor_4.x/Skin_SDK)
documentation.

## Installation

**Just skin please**

Add the whole bootstrapck folder to the skin folder.<br />
In ckeditor.js and config.js change the skin name to "bootstrapck".<br />
Done!

**The whole skin - sass - grunt package**

All the sass files are included in the bootstrapck folder, so first follow the 'just skin please'-steps<br />
Now add the Gruntfile.js and the package.json to de ckeditor folder.

    npm install
    grunt build

You can start tampering now.

## Demo

http://kunstmaan.github.io/BootstrapCK4-Skin/

### Previous version

If you would like to get the Bootstrap2 skin for CKeditor3, [here](https://github.com/Kunstmaan/BootstrapCK-Skin)'s the previous version.
