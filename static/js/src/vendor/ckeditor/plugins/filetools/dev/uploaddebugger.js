/**
 * @license Copyright (c) 2003-2017, CKSource - Frederico Knabben. All rights reserved.
 * For licensing, see LICENSE.md or http://ckeditor.com/license
 */

'use strict';

// Slow down the upload process.
// This trick works only on Chrome.

( function() {
	XMLHttpRequest.prototype.baseSend = XMLHttpRequest.prototype.send;

	XMLHttpRequest.prototype.send = function( data ) {
		var baseOnProgress = this.onprogress,
			baseOnLoad = this.onload;

		this.onprogress = function() {};

		this.onload = function( evt ) {
			// Total file size.
			var total = 1163,
				step = Math.round( total / 10 ),
				loaded = 0,
				xhr = this;

			function progress() {
				setTimeout( function() {
					if ( xhr.aborted ) {
						return;
					}

					loaded += step;
					if ( loaded > total ) {
						loaded = total;
					}

					if ( loaded > step * 4 && xhr.responseText.indexOf( 'incorrectFile' ) > 0 ) {
						xhr.aborted = true;
						xhr.onerror();
					} else if ( loaded < total ) {
						evt.loaded = loaded;
						baseOnProgress( { loaded: loaded } );
						progress();
					} else {
						baseOnLoad( evt );
					}
				}, 300 );
			}

			progress();
		};

		this.abort = function() {
			this.aborted = true;
			this.onabort();
		};

		this.baseSend( data );
	};
} )();
