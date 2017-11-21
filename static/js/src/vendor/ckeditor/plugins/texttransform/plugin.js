/**
 * @authors: Önder Ceylan <onderceylan@gmail.com>, PPKRAUSS https://github.com/ppKrauss
 * @license Licensed under the terms of GPL, LGPL and MPL licenses.
 * @version 1.1
 * @history v1.0 at 2013-05-09 by onderceylan, v1.1 at 2013-08-27 by ppkrauss.
 */

CKEDITOR.plugins.add('texttransform',
    {

        // define lang codes for available lang files here
        lang: 'en,tr',

        // plugin initialise
        init: function(editor)
        {
            // set num for switcher loop
            var num = 0;

            // add transformTextSwitch command to be used with button
            editor.addCommand('transformTextSwitch',
                {
                    exec : function()
                    {
                        var selection = editor.getSelection();
                        var commandArray = ['transformTextToUppercase', 'transformTextToLowercase', 'transformTextCapitalize'];

                        if (selection.getSelectedText().length > 0) {

                            selection.lock();

                            editor.execCommand( commandArray[num] );

                            selection.unlock(true);

                            if (num < commandArray.length - 1) {
                                num++;
                            } else {
                                num = 0;
                            }

                        }
                    }
                });

            // add transformTextToUppercase command to be used with buttons and 'execCommand' method
            editor.addCommand('transformTextToUppercase',
                {
                    exec : function()
                    {
                        var selection = editor.getSelection();
                        if (selection.getSelectedText().length > 0) {
				var ranges = selection.getRanges(),
				    walker = new CKEDITOR.dom.walker( ranges[0] ), 
				    node;
				while ( ( node = walker.next() ) ) 
					if ( node.type == CKEDITOR.NODE_TEXT && node.getText() ) 
						if (editor.langCode == "tr") {
							node.$.textContent = node.$.textContent.trToUpperCase();
						} else {
							node.$.textContent = node.$.textContent.toLocaleUpperCase();		
						}
                        }//if
                    } //func
                });

            // add transformTextToUppercase command to be used with buttons and 'execCommand' method
            editor.addCommand('transformTextToLowercase',
                {
                    exec : function()
                    {
                        var selection = editor.getSelection();
                        if (selection.getSelectedText().length > 0) {
				var ranges = selection.getRanges(),
				    walker = new CKEDITOR.dom.walker( ranges[0] ), 
				    node;
				while ( ( node = walker.next() ) ) 
					if ( node.type == CKEDITOR.NODE_TEXT && node.getText() ) 
						if (editor.langCode == "tr") {
							node.$.textContent = node.$.textContent.trToLowerCase();
						} else {
							node.$.textContent = node.$.textContent.toLocaleLowerCase();		
						}
						
                        }//if

                    }
                });

            // add transformTextCapitalize command to be used with buttons and 'execCommand' method
            editor.addCommand( 'transformTextCapitalize',
                {
                    exec : function()
                    {
                        var selection = editor.getSelection();
                        if (selection.getSelectedText().length > 0) {
				var ranges = selection.getRanges(),
				    walker = new CKEDITOR.dom.walker( ranges[0] ), 
				    node;
				while ( ( node = walker.next() ) ) 
					if ( node.type == CKEDITOR.NODE_TEXT && node.getText() )
						node.$.textContent = node.$.textContent.replace(
							/[^\s]\S*/g, 
							function(txt){
								if (editor.langCode == "tr") {
									return  txt.charAt(0).trToUpperCase() + 
										txt.substr(1).trToLowerCase();
								} else {
									return  txt.charAt(0).toLocaleUpperCase() + 
										txt.substr(1).toLocaleLowerCase();
								}
							
							}
						);
                        }//if
                    }
                });

            // add TransformTextSwitcher button to editor
            editor.ui.addButton('TransformTextSwitcher',
                {
                    label: editor.lang.texttransform.transformTextSwitchLabel,
                    command: 'transformTextSwitch',
                    icon: this.path + 'images/transformSwitcher.png'
                } );

            // add TransformTextToLowercase button to editor
            editor.ui.addButton('TransformTextToLowercase',
                {
                    label: editor.lang.texttransform.transformTextToLowercaseLabel,
                    command: 'transformTextToLowercase',
                    icon: this.path + 'images/transformToLower.png'
                } );

            // add TransformTextToUppercase button to editor
            editor.ui.addButton('TransformTextToUppercase',
                {
                    label: editor.lang.texttransform.transformTextToUppercaseLabel,
                    command: 'transformTextToUppercase',
                    icon: this.path + 'images/transformToUpper.png'
                } );

            // add TransformTextCapitalize button to editor
            editor.ui.addButton('TransformTextCapitalize',
                {
                    label: editor.lang.texttransform.transformTextCapitalizeLabel,
                    command: 'transformTextCapitalize',
                    icon: this.path + 'images/transformCapitalize.png'
                } );
        }
    } );
