Text Transform Plugin for CKEditor
================================

A very simple plugin which provides transforming selected text to new cases. You can transform selected text to uppercase, lowercase or simply capitalize it.

Available Transform Cases
-------------------------

* Transform Text to Uppercase: Convert letters to uppercase
* Transform Text to Lowercase: Convert letters to lowercase
* Transform Capitalize: Capitalize each word of selected text
* Transform Switcher: Loop through all cases

Internationalization
-------------------------

Currently plugin supports 2 languages.

* en
* tr

*Translations are welcomed.*

Usage
-------------------------

1. Download source files and place them on to be created "texttransform" folder under the CKEditor's plugin base.

2. Define plugin in CKEDITOR config object.

        CKEDITOR.config.extraPlugins = 'texttransform';

3. Add transform buttons to your editor toolbar.

        CKEDITOR.config.toolbar = [
            { name: 'texttransform', items: [ 'TransformTextToUppercase', 'TransformTextToLowercase', 'TransformTextCapitalize', 'TransformTextSwitcher' ] }
        ];

4. Set your CKEDITOR language if you did not set it yet.

        CKEDITOR.config.language = 'en';

Demo
-------------------------

[View the live demo](http://jsfiddle.net/t99kV/5/) on jsFiddle.


Cheers
--------------------

Thanks to [CKEditor] [1] and [jsFiddle] [2] for their good work.

  [1]: http://ckeditor.com        "CKEditor"
  [2]: http://jsfiddle.net        "jsFiddle"
