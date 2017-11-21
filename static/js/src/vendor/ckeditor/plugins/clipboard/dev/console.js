/**
 * @license Copyright (c) 2003-2017, CKSource - Frederico Knabben. All rights reserved.
 * For licensing, see LICENSE.md or http://ckeditor.com/license
 */

/* global CKCONSOLE */

'use strict';

( function() {
	var pasteType, pasteValue;

	CKCONSOLE.add( 'paste', {
		panels: [
			{
				type: 'box',
				content:
				'<ul class="ckconsole_list">' +
					'<li>type: <span class="ckconsole_value" data-value="type"></span></li>' +
					'<li>value: <span class="ckconsole_value" data-value="value"></span></li>' +
				'</ul>',

				refresh: function() {
					return {
						header: 'Paste',
						type: pasteType,
						value: pasteValue
					};
				},

				refreshOn: function( editor, refresh ) {
					editor.on( 'paste', function( evt ) {
						pasteType = evt.data.type;
						pasteValue = CKEDITOR.tools.htmlEncode( evt.data.dataValue );
						refresh();
					} );
				}
			},
			{
				type: 'log',
				on: function( editor, log, logFn ) {
					editor.on( 'paste', function( evt ) {
						logFn( 'paste; type:' + evt.data.type )();
					} );
				}
			}
		]
	} );
} )();
