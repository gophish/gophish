/*
 * @file Zoom plugin for CKEditor
 * Copyright (C) 2008-2013 Alfonso Martínez de Lizarrondo
 * Upgrade to CKEditor 4 sponsored by Solution architects gmbh
 *
 * == BEGIN LICENSE ==
 *
 * Licensed under the terms of any of the following licenses at your
 * choice:
 *
 *  - GNU General Public License Version 2 or later (the "GPL")
 *    http://www.gnu.org/licenses/gpl.html
 *
 *  - GNU Lesser General Public License Version 2.1 or later (the "LGPL")
 *    http://www.gnu.org/licenses/lgpl.html
 *
 *  - Mozilla Public License Version 1.1 or later (the "MPL")
 *    http://www.mozilla.org/MPL/MPL-1.1.html
 *
 * == END LICENSE ==
 *
 */

CKEDITOR.plugins.add( 'zoom',
{
	requires : [ 'richcombo' ],

	init : function( editor )
	{
		var config = editor.config;

		// Inject basic sizing for the pane as the richCombo doesn't allow to specify it
		var node = CKEDITOR.document.getHead().append( 'style' );
		node.setAttribute( 'type', 'text/css' );
		var content = '.cke_combopanel__zoom { height: 200px; width: 100px; }' +
					'.cke_combo__zoom .cke_combo_text { width: 40px;}';

		if ( CKEDITOR.env.ie && CKEDITOR.env.version<11 )
			node.$.styleSheet.cssText = content;
		else
			node.$.innerHTML = content;

		editor.ui.addRichCombo( 'Zoom',
			{
				label : 'Zoom',
				title : 'Zoom',
				multiSelect : false,
				className : 'zoom',
				modes:{wysiwyg:1,source:1 },

				panel :
				{
					css : [ CKEDITOR.skin.getPath( 'editor' ) ].concat( config.contentsCss )
				},

				init : function()
				{
					var zoomOptions = [50, 75, 100, 125, 150, 200, 400],
						zoom;

					this.startGroup( 'Zoom level' );
					// Loop over the Array, adding all items to the combo.
					for ( var i = 0 ; i < zoomOptions.length ; i++ )
					{
						zoom = zoomOptions[ i ];
						// value, html, text
						this.add( zoom + "", zoom + " %", zoom + " %" );
					}
					// Default value on first click
					this.setValue("100", "100 %");
				},

				onClick : function( sValue )
				{
					var body = editor.editable().$;
					var value = parseInt(sValue);

					body.style.MozTransformOrigin = "top left";
					body.style.MozTransform = "scale(" + (value/100)  + ")";

					body.style.WebkitTransformOrigin = "top left";
					body.style.WebkitTransform = "scale(" + (value/100)  + ")";

					body.style.OTransformOrigin = "top left";
					body.style.OTransform = "scale(" + (value/100)  + ")";

					body.style.TransformOrigin = "top left";
					body.style.Transform = "scale(" + (value/100)  + ")";
					// IE
					body.style.zoom = value/100;

					this.setValue( sValue, sValue + " %");
					this.lastValue = sValue;
				},

				onRender: function() {
					editor.on( 'mode', function( ev ) {
						// Restore zoom level after switching from Source mode
						if (this.lastValue)
							this.onClick( this.lastValue );

					}, this );
				}
			});
		// End of richCombo element

	} //Init
} );

