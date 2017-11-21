# Emojione for CKEditor

This plugin integrates the emojione library into the CKEditor. The plugin allows you to add all known emojis into your content in unicode format.

Plugin on the [CKEditor website](http://ckeditor.com/addon/emojione)

Try our [Demo](http://ckeditor-emojione-demo.braune-digital.com)

You can install the dependencies and the plugin with bower:

```
bower install ckeditor
bower install emojione
bower install ckeditor-emojione
```

After that you can register the plugin within your CKeditor configuration:

```
<!DOCTYPE html>
<html>
    <head>
        <meta charset="utf-8">
        <title>CKEditor</title>
        <script src="bower_components/emojione/lib/js/emojione.js"></script>
        <script src="bower_components/ckeditor/ckeditor.js"></script>
    </head>
    <body>
        <form>
            <textarea name="editor" id="editor" rows="10" cols="80">
                This is my textarea to be replaced with CKEditor.
            </textarea>
            <script>
                CKEDITOR.plugins.addExternal('emojione', '../../bower_components/ckeditor-emojione/', 'plugin.js');
                CKEDITOR.config.extraPlugins = 'emojione';
                CKEDITOR.replace( 'editor' );
            </script>
        </form>
    </body>
</html>
```

This addon has been inspired by the [smiley plugin](https://github.com/ckeditor/ckeditor-dev/tree/master/plugins/smiley) and added support for native emojis.

