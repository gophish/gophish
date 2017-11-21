/**
 * @license Copyright (c) 2003-2017, CKSource - Frederico Knabben. All rights reserved.
 * For licensing, see LICENSE.md or http://ckeditor.com/license
 */

CKEDITOR.plugins.add( 'docprops', {
	requires: 'wysiwygarea,dialog,colordialog',
	// jscs:disable maximumLineLength
	lang: 'af,ar,az,bg,bn,bs,ca,cs,cy,da,de,de-ch,el,en,en-au,en-ca,en-gb,eo,es,es-mx,et,eu,fa,fi,fo,fr,fr-ca,gl,gu,he,hi,hr,hu,id,is,it,ja,ka,km,ko,ku,lt,lv,mk,mn,ms,nb,nl,no,oc,pl,pt,pt-br,ro,ru,si,sk,sl,sq,sr,sr-latn,sv,th,tr,tt,ug,uk,vi,zh,zh-cn', // %REMOVE_LINE_CORE%
	// jscs:enable maximumLineLength
	icons: 'docprops,docprops-rtl', // %REMOVE_LINE_CORE%
	hidpi: true, // %REMOVE_LINE_CORE%
	init: function( editor ) {
		var cmd = new CKEDITOR.dialogCommand( 'docProps' );
		// Only applicable on full page mode.
		cmd.modes = { wysiwyg: editor.config.fullPage };
		cmd.allowedContent = {
			body: {
				styles: '*',
				attributes: 'dir'
			},
			html: {
				attributes: 'lang,xml:lang'
			}
		};
		cmd.requiredContent = 'body';

		editor.addCommand( 'docProps', cmd );
		CKEDITOR.dialog.add( 'docProps', this.path + 'dialogs/docprops.js' );

		editor.ui.addButton && editor.ui.addButton( 'DocProps', {
			label: editor.lang.docprops.label,
			command: 'docProps',
			toolbar: 'document,30'
		} );
	}
} );
