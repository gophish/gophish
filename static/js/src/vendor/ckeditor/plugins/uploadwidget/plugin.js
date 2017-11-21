/**
 * @license Copyright (c) 2003-2017, CKSource - Frederico Knabben. All rights reserved.
 * For licensing, see LICENSE.md or http://ckeditor.com/license
 */

'use strict';

( function() {
	CKEDITOR.plugins.add( 'uploadwidget', {
		lang: 'az,ca,cs,da,de,de-ch,el,en,eo,es,es-mx,eu,fr,gl,hr,hu,id,it,ja,km,ko,ku,nb,nl,no,oc,pl,pt,pt-br,ru,sk,sv,tr,ug,uk,zh,zh-cn', // %REMOVE_LINE_CORE%
		requires: 'widget,clipboard,filetools,notificationaggregator',

		init: function( editor ) {
			// Images which should be changed into upload widget needs to be marked with `data-widget` on paste,
			// because otherwise wrong widget may handle upload placeholder element (e.g. image2 plugin would handle image).
			// `data-widget` attribute is allowed only in the elements which has also `data-cke-upload-id` attribute.
			editor.filter.allow( '*[!data-widget,!data-cke-upload-id]' );
		}
	} );

	/**
	 * This function creates an upload widget &mdash; a placeholder to show the progress of an upload. The upload widget
	 * is based on its {@link CKEDITOR.fileTools.uploadWidgetDefinition definition}. The `addUploadWidget` method also
	 * creates a `paste` event, if the {@link CKEDITOR.fileTools.uploadWidgetDefinition#fileToElement fileToElement} method
	 * is defined. This event helps in handling pasted files, as it will automatically check if the files were pasted and
	 * mark them to be uploaded.
	 *
	 * The upload widget helps to handle content that is uploaded asynchronously inside the editor. It solves issues such as:
	 * editing during upload, undo manager integration, getting data, removing or copying uploaded element.
	 *
	 * To create an upload widget you need to define two transformation methods:
	 *
	 * * The {@link CKEDITOR.fileTools.uploadWidgetDefinition#fileToElement fileToElement} method which will be called on
	 * `paste` and transform a file into a placeholder.
	 * * The {@link CKEDITOR.fileTools.uploadWidgetDefinition#onUploaded onUploaded} method with
	 * the {@link CKEDITOR.fileTools.uploadWidgetDefinition#replaceWith replaceWith} method which will be called to replace
	 * the upload placeholder with the final HTML when the upload is done.
	 * If you want to show more information about the progress you can also define
	 * the {@link CKEDITOR.fileTools.uploadWidgetDefinition#onLoading onLoading} and
	 * {@link CKEDITOR.fileTools.uploadWidgetDefinition#onUploading onUploading} methods.
	 *
	 * The simplest uploading widget which uploads a file and creates a link to it may look like this:
	 *
	 * 		CKEDITOR.fileTools.addUploadWidget( editor, 'uploadfile', {
	 *			uploadUrl: CKEDITOR.fileTools.getUploadUrl( editor.config ),
	 *
	 *			fileToElement: function( file ) {
	 *				var a = new CKEDITOR.dom.element( 'a' );
	 *				a.setText( file.name );
	 *				a.setAttribute( 'href', '#' );
	 *				return a;
	 *			},
	 *
	 *			onUploaded: function( upload ) {
	 *				this.replaceWith( '<a href="' + upload.url + '" target="_blank">' + upload.fileName + '</a>' );
	 *			}
	 *		} );
	 *
	 * The upload widget uses {@link CKEDITOR.fileTools.fileLoader} as a helper to upload the file. A
	 * {@link CKEDITOR.fileTools.fileLoader} instance is created when the file is pasted and a proper method is
	 * called &mdash; by default it is the {@link CKEDITOR.fileTools.fileLoader#loadAndUpload} method. If you want
	 * to only use the `load` or `upload`, you can use the {@link CKEDITOR.fileTools.uploadWidgetDefinition#loadMethod loadMethod}
	 * property.
	 *
	 * Note that if you want to handle a big file, e.g. a video, you may need to use `upload` instead of
	 * `loadAndUpload` because the file may be too big to load it to memory at once.
	 *
	 * If you do not upload the file, you need to define {@link CKEDITOR.fileTools.uploadWidgetDefinition#onLoaded onLoaded}
	 * instead of {@link CKEDITOR.fileTools.uploadWidgetDefinition#onUploaded onUploaded}.
	 * For example, if you want to read the content of the file:
	 *
	 *		CKEDITOR.fileTools.addUploadWidget( editor, 'fileReader', {
	 *			loadMethod: 'load',
	 *			supportedTypes: /text\/(plain|html)/,
	 *
	 *			fileToElement: function( file ) {
	 *				var el = new CKEDITOR.dom.element( 'span' );
	 *				el.setText( '...' );
	 *				return el;
	 *			},
	 *
	 *			onLoaded: function( loader ) {
	 *				this.replaceWith( atob( loader.data.split( ',' )[ 1 ] ) );
	 *			}
	 *		} );
	 *
	 * If you need to pass additional data to the request, you can do this using the
	 * {@link CKEDITOR.fileTools.uploadWidgetDefinition#additionalRequestParameters additionalRequestParameters} property.
	 * That data is then passed to the upload method defined by {@link CKEDITOR.fileTools.uploadWidgetDefinition#loadMethod},
	 * and to the {@link CKEDITOR.editor#fileUploadRequest} event (as part of the `requestData` property).
	 * The syntax of that parameter is compatible with the {@link CKEDITOR.editor#fileUploadRequest} `requestData` property.
	 *
	 *		CKEDITOR.fileTools.addUploadWidget( editor, 'uploadFile', {
	 *			additionalRequestParameters: {
	 *				foo: 'bar'
	 *			},
	 *
	 *			fileToElement: function( file ) {
	 *				var el = new CKEDITOR.dom.element( 'span' );
	 *				el.setText( '...' );
	 *				return el;
	 *			},
	 *
	 *			onUploaded: function( upload ) {
	 *				this.replaceWith( '<a href="' + upload.url + '" target="_blank">' + upload.fileName + '</a>' );
	 *			}
	 *		} );
	 *
	 * If you need custom `paste` handling, you need to mark the pasted element to be changed into an upload widget
	 * using {@link CKEDITOR.fileTools#markElement markElement}. For example, instead of the `fileToElement` helper from the
	 * example above, a `paste` listener can be created manually:
	 *
	 *		editor.on( 'paste', function( evt ) {
	 *			var file, i, el, loader;
	 *
	 *			for ( i = 0; i < evt.data.dataTransfer.getFilesCount(); i++ ) {
	 *				file = evt.data.dataTransfer.getFile( i );
	 *
	 *				if ( CKEDITOR.fileTools.isTypeSupported( file, /text\/(plain|html)/ ) ) {
	 *					el = new CKEDITOR.dom.element( 'span' ),
	 *					loader = editor.uploadRepository.create( file );
	 *
	 *					el.setText( '...' );
	 *
	 *					loader.load();
	 *
	 *					CKEDITOR.fileTools.markElement( el, 'filereader', loader.id );
	 *
	 *					evt.data.dataValue += el.getOuterHtml();
	 *				}
	 *			}
	 *		} );
	 *
	 * Note that you can bind notifications to the upload widget on paste using
	 * the {@link CKEDITOR.fileTools#bindNotifications} method, so notifications will automatically
	 * show the progress of the upload. Because this method shows notifications about upload, do not use it if you only
	 * {@link CKEDITOR.fileTools.fileLoader#load load} (and not upload) a file.
	 *
	 *		editor.on( 'paste', function( evt ) {
	 *			var file, i, el, loader;
	 *
	 *			for ( i = 0; i < evt.data.dataTransfer.getFilesCount(); i++ ) {
	 *				file = evt.data.dataTransfer.getFile( i );
	 *
	 *				if ( CKEDITOR.fileTools.isTypeSupported( file, /text\/pdf/ ) ) {
	 *					el = new CKEDITOR.dom.element( 'span' ),
	 *					loader = editor.uploadRepository.create( file );
	 *
	 *					el.setText( '...' );
	 *
	 *					loader.upload();
	 *
	 *					CKEDITOR.fileTools.markElement( el, 'pdfuploader', loader.id );
	 *
	 *					CKEDITOR.fileTools.bindNotifications( editor, loader );
	 *
	 *					evt.data.dataValue += el.getOuterHtml();
	 *				}
	 *			}
	 *		} );
	 *
	 * @member CKEDITOR.fileTools
	 * @param {CKEDITOR.editor} editor The editor instance.
	 * @param {String} name The name of the upload widget.
	 * @param {CKEDITOR.fileTools.uploadWidgetDefinition} def Upload widget definition.
	 */
	function addUploadWidget( editor, name, def ) {
		var fileTools = CKEDITOR.fileTools,
			uploads = editor.uploadRepository,
			// Plugins which support all file type has lower priority than plugins which support specific types.
			priority = def.supportedTypes ? 10 : 20;

		if ( def.fileToElement ) {
			editor.on( 'paste', function( evt ) {
				var data = evt.data,
					dataTransfer = data.dataTransfer,
					filesCount = dataTransfer.getFilesCount(),
					loadMethod = def.loadMethod || 'loadAndUpload',
					file, i;

				if ( data.dataValue || !filesCount ) {
					return;
				}

				for ( i = 0; i < filesCount; i++ ) {
					file = dataTransfer.getFile( i );

					// No def.supportedTypes means all types are supported.
					if ( !def.supportedTypes || fileTools.isTypeSupported( file, def.supportedTypes ) ) {
						var el = def.fileToElement( file ),
							loader = uploads.create( file );

						if ( el ) {
							loader[ loadMethod ]( def.uploadUrl, def.additionalRequestParameters );

							CKEDITOR.fileTools.markElement( el, name, loader.id );

							if ( loadMethod == 'loadAndUpload' || loadMethod == 'upload' ) {
								CKEDITOR.fileTools.bindNotifications( editor, loader );
							}

							data.dataValue += el.getOuterHtml();
						}
					}
				}
			}, null, null, priority );
		}

		/**
		 * This is an abstract class that describes a definition of an upload widget.
		 * It is a type of {@link CKEDITOR.fileTools#addUploadWidget} method's second argument.
		 *
		 * Note that because the upload widget is a type of a widget, this definition extends
		 * {@link CKEDITOR.plugins.widget.definition}.
		 * It adds several new properties and callbacks and implements the {@link CKEDITOR.plugins.widget.definition#downcast}
		 * and {@link CKEDITOR.plugins.widget.definition#init} callbacks. These two properties
		 * should not be overwritten.
		 *
		 * Also, the upload widget definition defines a few properties ({@link #fileToElement}, {@link #supportedTypes},
		 * {@link #loadMethod loadMethod}, {@link #uploadUrl} and {@link #additionalRequestParameters}) used in the
		 * {@link CKEDITOR.editor#paste} listener which is registered by {@link CKEDITOR.fileTools#addUploadWidget}
		 * if the upload widget definition contains the {@link #fileToElement} callback.
		 *
		 * @abstract
		 * @class CKEDITOR.fileTools.uploadWidgetDefinition
		 * @mixins CKEDITOR.plugins.widget.definition
		 */
		CKEDITOR.tools.extend( def, {
			/**
			 * Upload widget definition overwrites the {@link CKEDITOR.plugins.widget.definition#downcast} property.
			 * This should not be changed.
			 *
			 * @property {String/Function}
			 */
			downcast: function() {
				return new CKEDITOR.htmlParser.text( '' );
			},

			/**
			 * Upload widget definition overwrites the {@link CKEDITOR.plugins.widget.definition#init} property.
			 * If you want to add some code in the `init` callback remember to call the base function.
			 *
			 * @property {Function}
			 */
			init: function() {
				var widget = this,
					id = this.wrapper.findOne( '[data-cke-upload-id]' ).data( 'cke-upload-id' ),
					loader = uploads.loaders[ id ],
					capitalize = CKEDITOR.tools.capitalize,
					oldStyle, newStyle;

				loader.on( 'update', function( evt ) {
					// Abort if widget was removed.
					if ( !widget.wrapper || !widget.wrapper.getParent() ) {
						if ( !editor.editable().find( '[data-cke-upload-id="' + id + '"]' ).count() ) {
							loader.abort();
						}
						evt.removeListener();
						return;
					}

					editor.fire( 'lockSnapshot' );

					// Call users method, eg. if the status is `uploaded` then
					// `onUploaded` method will be called, if exists.
					var methodName = 'on' + capitalize( loader.status );

					if ( typeof widget[ methodName ] === 'function' ) {
						if ( widget[ methodName ]( loader ) === false ) {
							editor.fire( 'unlockSnapshot' );
							return;
						}
					}

					// Set style to the wrapper if it still exists.
					newStyle = 'cke_upload_' + loader.status;
					if ( widget.wrapper && newStyle != oldStyle ) {
						oldStyle && widget.wrapper.removeClass( oldStyle );
						widget.wrapper.addClass( newStyle );
						oldStyle = newStyle;
					}

					// Remove widget on error or abort.
					if ( loader.status == 'error' || loader.status == 'abort' ) {
						editor.widgets.del( widget );
					}

					editor.fire( 'unlockSnapshot' );
				} );

				loader.update();
			},

			/**
			 * Replaces the upload widget with the final HTML. This method should be called when the upload is done,
			 * usually in the {@link #onUploaded} callback.
			 *
			 * @property {Function}
			 * @param {String} data HTML to replace the upload widget.
			 * @param {String} [mode='html'] See {@link CKEDITOR.editor#method-insertHtml}'s modes.
			 */
			replaceWith: function( data, mode ) {
				if ( data.trim() === '' ) {
					editor.widgets.del( this );
					return;
				}

				var wasSelected = ( this == editor.widgets.focused ),
					editable = editor.editable(),
					range = editor.createRange(),
					bookmark, bookmarks;

				if ( !wasSelected ) {
					bookmarks = editor.getSelection().createBookmarks();
				}

				range.setStartBefore( this.wrapper );
				range.setEndAfter( this.wrapper );

				if ( wasSelected ) {
					bookmark = range.createBookmark();
				}

				editable.insertHtmlIntoRange( data, range, mode );

				editor.widgets.checkWidgets( { initOnlyNew: true } );

				// Ensure that old widgets instance will be removed.
				// If replaceWith is called in init, because of paste then checkWidgets will not remove it.
				editor.widgets.destroy( this, true );

				if ( wasSelected ) {
					range.moveToBookmark( bookmark );
					range.select();
				} else {
					editor.getSelection().selectBookmarks( bookmarks );
				}

			}

			/**
			 * If this property is defined, paste listener is created to transform the pasted file into an HTML element.
			 * It creates an HTML element which will be then transformed into an upload widget.
			 * It is only called for {@link #supportedTypes supported files}.
			 * If multiple files were pasted, this function will be called for each file of a supported type.
			 *
			 * @property {Function} fileToElement
			 * @param {Blob} file A pasted file to load or upload.
			 * @returns {CKEDITOR.dom.element} An element which will be transformed into the upload widget.
			 */

			/**
			 * Regular expression to check if the file type is supported by this widget.
			 * If not defined, all files will be handled.
			 *
			 * @property {String} [supportedTypes]
			 */

			/**
			 * The URL to which the file will be uploaded. It should be taken from the configuration using
			 * {@link CKEDITOR.fileTools#getUploadUrl}.
			 *
			 * @property {String} [uploadUrl]
			 */

			/**
			 * An object containing additional data that should be passed to the function defined by {@link #loadMethod}.
			 *
			 * @property {Object} [additionalRequestParameters]
			 */

			/**
			 * The type of loading operation that should be executed as a result of pasting a file. Possible options are:
			 *
			 * * `'loadAndUpload'` &ndash; Default behavior. The {@link CKEDITOR.fileTools.fileLoader#loadAndUpload} method will be
			 * executed, the file will be loaded first and uploaded immediately after loading is done.
			 * * `'load'` &ndash; The {@link CKEDITOR.fileTools.fileLoader#load} method will be executed. This loading type should
			 * be used if you only want to load file data without uploading it.
			 * * `'upload'` &ndash; The {@link CKEDITOR.fileTools.fileLoader#upload} method will be executed, the file will be uploaded
			 * without loading it to memory. This loading type should be used if you want to upload a big file,
			 * otherwise you can get an "out of memory" error.
			 *
			 * @property {String} [loadMethod=loadAndUpload]
			 */

			/**
			 * A function called when the {@link CKEDITOR.fileTools.fileLoader#status status} of the upload changes to `loading`.
			 *
			 * @property {Function} [onLoading]
			 * @param {CKEDITOR.fileTools.fileLoader} loader Loader instance.
			 * @returns {Boolean}
			 */

			/**
			 * A function called when the {@link CKEDITOR.fileTools.fileLoader#status status} of the upload changes to `loaded`.
			 *
			 * @property {Function} [onLoaded]
			 * @param {CKEDITOR.fileTools.fileLoader} loader Loader instance.
			 * @returns {Boolean}
			 */

			/**
			 * A function called when the {@link CKEDITOR.fileTools.fileLoader#status status} of the upload changes to `uploading`.
			 *
			 * @property {Function} [onUploading]
			 * @param {CKEDITOR.fileTools.fileLoader} loader Loader instance.
			 * @returns {Boolean}
			 */

			/**
			 * A function called when the {@link CKEDITOR.fileTools.fileLoader#status status} of the upload changes to `uploaded`.
			 * At that point the upload is done and the upload widget should be replaced with the final HTML using
			 * the {@link #replaceWith} method.
			 *
			 * @property {Function} [onUploaded]
			 * @param {CKEDITOR.fileTools.fileLoader} loader Loader instance.
			 * @returns {Boolean}
			 */

			/**
			 * A function called when the {@link CKEDITOR.fileTools.fileLoader#status status} of the upload changes to `error`.
			 * The default behavior is to remove the widget and it can be canceled if this function returns `false`.
			 *
			 * @property {Function} [onError]
			 * @param {CKEDITOR.fileTools.fileLoader} loader Loader instance.
			 * @returns {Boolean} If `false`, the default behavior (remove widget) will be canceled.
			 */

			/**
			 * A function called when the {@link CKEDITOR.fileTools.fileLoader#status status} of the upload changes to `abort`.
			 * The default behavior is to remove the widget and it can be canceled if this function returns `false`.
			 *
			 * @property {Function} [onAbort]
			 * @param {CKEDITOR.fileTools.fileLoader} loader Loader instance.
			 * @returns {Boolean} If `false`, the default behavior (remove widget) will be canceled.
			 */
		} );

		editor.widgets.add( name, def );
	}

	/**
	 * Marks an element which should be transformed into an upload widget.
	 *
	 * @see CKEDITOR.fileTools.addUploadWidget
	 *
	 * @member CKEDITOR.fileTools
	 * @param {CKEDITOR.dom.element} element Element to be marked.
	 * @param {String} widgetName The name of the upload widget.
	 * @param {Number} loaderId The ID of a related {@link CKEDITOR.fileTools.fileLoader}.
	 */
	function markElement( element, widgetName, loaderId  ) {
		element.setAttributes( {
			'data-cke-upload-id': loaderId,
			'data-widget': widgetName
		} );
	}

	/**
	 * Binds a notification to the {@link CKEDITOR.fileTools.fileLoader file loader} so the upload widget will use
	 * the notification to show the status and progress.
	 * This function uses {@link CKEDITOR.plugins.notificationAggregator}, so even if multiple files are uploading
	 * only one notification is shown. Warnings are an exception, because they are shown in separate notifications.
	 * This notification shows only the progress of the upload, so this method should not be used if
	 * the {@link CKEDITOR.fileTools.fileLoader#load loader.load} method was called. It works with
	 * {@link CKEDITOR.fileTools.fileLoader#upload upload} and {@link CKEDITOR.fileTools.fileLoader#loadAndUpload loadAndUpload}.
	 *
	 * @member CKEDITOR.fileTools
	 * @param {CKEDITOR.editor} editor The editor instance.
	 * @param {CKEDITOR.fileTools.fileLoader} loader The file loader instance.
	 */
	function bindNotifications( editor, loader ) {
		var aggregator,
			task = null;

		loader.on( 'update', function() {
			// Value of uploadTotal is known after upload start. Task will be created when uploadTotal is present.
			if ( !task && loader.uploadTotal ) {
				createAggregator();
				task = aggregator.createTask( { weight: loader.uploadTotal } );
			}

			if ( task && loader.status == 'uploading' ) {
				task.update( loader.uploaded );
			}
		} );

		loader.on( 'uploaded', function() {
			task && task.done();
		} );

		loader.on( 'error', function() {
			task && task.cancel();
			editor.showNotification( loader.message, 'warning' );
		} );

		loader.on( 'abort', function() {
			task && task.cancel();
			editor.showNotification( editor.lang.uploadwidget.abort, 'info' );
		} );

		function createAggregator() {
			aggregator = editor._.uploadWidgetNotificaionAggregator;

			// Create one notification aggregator for all types of upload widgets for the editor.
			if ( !aggregator || aggregator.isFinished() ) {
				aggregator = editor._.uploadWidgetNotificaionAggregator = new CKEDITOR.plugins.notificationAggregator(
					editor, editor.lang.uploadwidget.uploadMany, editor.lang.uploadwidget.uploadOne );

				aggregator.once( 'finished', function() {
					var tasks = aggregator.getTaskCount();

					if ( tasks === 0 ) {
						aggregator.notification.hide();
					} else {
						aggregator.notification.update( {
							message: tasks == 1 ?
								editor.lang.uploadwidget.doneOne :
								editor.lang.uploadwidget.doneMany.replace( '%1', tasks ),
							type: 'success',
							important: 1
						} );
					}
				} );
			}
		}
	}

	// Two plugins extend this object.
	if ( !CKEDITOR.fileTools ) {
		CKEDITOR.fileTools = {};
	}

	CKEDITOR.tools.extend( CKEDITOR.fileTools, {
		addUploadWidget: addUploadWidget,
		markElement: markElement,
		bindNotifications: bindNotifications
	} );
} )();
