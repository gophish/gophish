/**
 * @license Copyright (c) 2003-2017, CKSource - Frederico Knabben. All rights reserved.
 * For licensing, see LICENSE.md or http://ckeditor.com/license
 */

/**
 * @fileOverview The [Language](http://ckeditor.com/addon/language) plugin.
 */

'use strict';

( function() {

	var allowedContent = 'span[!lang,!dir]',
		requiredContent = 'span[lang,dir]';

	CKEDITOR.plugins.add( 'language', {
		requires: 'menubutton',
		lang: 'ar,az,bg,ca,cs,cy,da,de,de-ch,el,en,en-gb,eo,es,es-mx,eu,fa,fi,fo,fr,gl,he,hr,hu,id,it,ja,km,ko,ku,nb,nl,no,oc,pl,pt,pt-br,ru,sk,sl,sq,sv,tr,tt,ug,uk,vi,zh,zh-cn', // %REMOVE_LINE_CORE%
		icons: 'language', // %REMOVE_LINE_CORE%
		hidpi: true, // %REMOVE_LINE_CORE%

		init: function( editor ) {
			var languagesConfigStrings = ( editor.config.language_list || [ 'ar:Arabic:rtl', 'fr:French', 'es:Spanish' ] ),
				plugin = this,
				lang = editor.lang.language,
				items = {},
				parts,
				curLanguageId, // 2-letter language identifier.
				languageButtonId, // Will store button namespaced identifier, like "language_en".
				i;

			// Registers command.
			editor.addCommand( 'language', {
				allowedContent: allowedContent,
				requiredContent: requiredContent,
				contextSensitive: true,
				exec: function( editor, languageId ) {
					var item = items[ 'language_' + languageId ];

					if ( item )
						editor[ item.style.checkActive( editor.elementPath(), editor ) ? 'removeStyle' : 'applyStyle' ]( item.style );
				},
				refresh: function( editor ) {
					this.setState( plugin.getCurrentLangElement( editor ) ?
						CKEDITOR.TRISTATE_ON : CKEDITOR.TRISTATE_OFF );
				}
			} );

			// Parse languagesConfigStrings, and create items entry for each lang.
			for ( i = 0; i < languagesConfigStrings.length; i++ ) {
				parts = languagesConfigStrings[ i ].split( ':' );
				curLanguageId = parts[ 0 ];
				languageButtonId = 'language_' + curLanguageId;

				items[ languageButtonId ] = {
					label: parts[ 1 ],
					langId: curLanguageId,
					group: 'language',
					order: i,
					// Tells if this language is left-to-right oriented (default: true).
					ltr: ( '' + parts[ 2 ] ).toLowerCase() != 'rtl',
					onClick: function() {
						editor.execCommand( 'language', this.langId );
					},
					role: 'menuitemcheckbox'
				};

				// Init style property.
				items[ languageButtonId ].style = new CKEDITOR.style( {
					element: 'span',
					attributes: {
						lang: curLanguageId,
						dir: items[ languageButtonId ].ltr ? 'ltr' : 'rtl'
					}
				} );
			}

			// Remove language indicator button.
			items.language_remove = {
				label: lang.remove,
				group: 'language_remove',
				state: CKEDITOR.TRISTATE_DISABLED,
				order: items.length,
				onClick: function() {
					var currentLanguagedElement = plugin.getCurrentLangElement( editor );

					if ( currentLanguagedElement )
						editor.execCommand( 'language', currentLanguagedElement.getAttribute( 'lang' ) );
				}
			};

			// Initialize groups for menu.
			editor.addMenuGroup( 'language', 1 );
			editor.addMenuGroup( 'language_remove' ); // Group order is skipped intentionally, it will be placed at the end.
			editor.addMenuItems( items );

			editor.ui.add( 'Language', CKEDITOR.UI_MENUBUTTON, {
				label: lang.button,
				// MenuButtons do not (yet) has toFeature method, so we cannot do this:
				// toFeature: function( editor ) { return editor.getCommand( 'language' ); }
				// Set feature's properties directly on button.
				allowedContent: allowedContent,
				requiredContent: requiredContent,
				toolbar: 'bidi,30',
				command: 'language',
				onMenu: function() {
					var activeItems = {},
						currentLanguagedElement = plugin.getCurrentLangElement( editor );

					for ( var prop in items )
						activeItems[ prop ] = CKEDITOR.TRISTATE_OFF;

					activeItems.language_remove = currentLanguagedElement ? CKEDITOR.TRISTATE_OFF : CKEDITOR.TRISTATE_DISABLED;

					if ( currentLanguagedElement )
						activeItems[ 'language_' + currentLanguagedElement.getAttribute( 'lang' ) ] = CKEDITOR.TRISTATE_ON;

					return activeItems;
				}
			} );
		},

		// Gets the first language element for the current editor selection.
		// @param {CKEDITOR.editor} editor
		// @returns {CKEDITOR.dom.element} The language element, if any.
		getCurrentLangElement: function( editor ) {
			var elementPath = editor.elementPath(),
				activePath = elementPath && elementPath.elements,
				pathMember, ret;

			// IE8: upon initialization if there is no path elementPath() returns null.
			if ( elementPath ) {
				for ( var i = 0; i < activePath.length; i++ ) {
					pathMember = activePath[ i ];

					if ( !ret && pathMember.getName() == 'span' && pathMember.hasAttribute( 'dir' ) && pathMember.hasAttribute( 'lang' ) )
						ret = pathMember;
				}
			}

			return ret;
		}
	} );
} )();

/**
 * Specifies the list of languages available in the
 * [Language](http://ckeditor.com/addon/language) plugin. Each entry
 * should be a string in the following format:
 *
 *		<languageCode>:<languageLabel>[:<textDirection>]
 *
 * * _languageCode_: The language code used for the `lang` attribute in ISO 639 format.
 * 	Language codes can be found [here](http://www.loc.gov/standards/iso639-2/php/English_list.php).
 * 	You can use both 2-letter ISO-639-1 codes and 3-letter ISO-639-2 codes, though
 * 	for consistency it is recommended to stick to ISO-639-1 2-letter codes.
 * * _languageLabel_: The label to show for this language in the list.
 * * _textDirection_: (optional) One of the following values: `rtl` or `ltr`,
 * 	indicating the reading direction of the language. Defaults to `ltr`.
 *
 * See the [SDK sample](http://sdk.ckeditor.com/samples/language.html).
 *
 *		config.language_list = [ 'he:Hebrew:rtl', 'pt:Portuguese', 'de:German' ];
 *
 * @cfg {Array} [language_list = [ 'ar:Arabic:rtl', 'fr:French', 'es:Spanish' ]]
 * @member CKEDITOR.config
 */
