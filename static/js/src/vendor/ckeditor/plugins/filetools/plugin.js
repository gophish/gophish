/**
 * @license Copyright (c) 2003-2017, CKSource - Frederico Knabben. All rights reserved.
 * For licensing, see LICENSE.md or http://ckeditor.com/license
 */

'use strict';

( function() {
	CKEDITOR.plugins.add( 'filetools', {
		lang: 'az,ca,cs,da,de,de-ch,en,eo,es,es-mx,eu,fr,gl,hr,hu,id,it,ja,km,ko,ku,nb,nl,oc,pl,pt,pt-br,ru,sk,sv,tr,ug,uk,zh,zh-cn', // %REMOVE_LINE_CORE%

		beforeInit: function( editor ) {
			/**
			 * An instance of the {@link CKEDITOR.fileTools.uploadRepository upload repository}.
			 * It allows you to create and get {@link CKEDITOR.fileTools.fileLoader file loaders}.
			 *
			 *		var loader = editor.uploadRepository.create( file );
			 *		loader.loadAndUpload( 'http://foo/bar' );
			 *
			 * @since 4.5
			 * @readonly
			 * @property {CKEDITOR.fileTools.uploadRepository} uploadRepository
			 * @member CKEDITOR.editor
			 */
			editor.uploadRepository = new UploadRepository( editor );

			/**
			 * Event fired when the {@link CKEDITOR.fileTools.fileLoader file loader} should send XHR. If the event is not
			 * {@link CKEDITOR.eventInfo#stop stopped} or {@link CKEDITOR.eventInfo#cancel canceled}, the default request
			 * will be sent. Refer to the [Uploading Dropped or Pasted Files](#!/guide/dev_file_upload) article for more information.
			 *
			 * @since 4.5
			 * @event fileUploadRequest
			 * @member CKEDITOR.editor
			 * @param data
			 * @param {CKEDITOR.fileTools.fileLoader} data.fileLoader A file loader instance.
			 * @param {Object} data.requestData An object containing all data to be sent to the server.
			 */
			editor.on( 'fileUploadRequest', function( evt ) {
				var fileLoader = evt.data.fileLoader;

				fileLoader.xhr.open( 'POST', fileLoader.uploadUrl, true );

				// Adding file to event's data by default - allows overwriting it by user's event listeners. (http://dev.ckeditor.com/ticket/13518)
				evt.data.requestData.upload = { file: fileLoader.file, name: fileLoader.fileName };
			}, null, null, 5 );

			editor.on( 'fileUploadRequest', function( evt ) {
				var fileLoader = evt.data.fileLoader,
					$formData = new FormData(),
					requestData = evt.data.requestData;

				for ( var name in requestData ) {
					var value = requestData[ name ];

					// Treating files in special way
					if ( typeof value === 'object' && value.file ) {
						$formData.append( name, value.file, value.name );
					}
					else {
						$formData.append( name, value );
					}
				}
				// Append token preventing CSRF attacks.
				$formData.append( 'ckCsrfToken', CKEDITOR.tools.getCsrfToken() );

				fileLoader.xhr.send( $formData );
			}, null, null, 999 );

			/**
			 * Event fired when the {CKEDITOR.fileTools.fileLoader file upload} response is received and needs to be parsed.
			 * If the event is not {@link CKEDITOR.eventInfo#stop stopped} or {@link CKEDITOR.eventInfo#cancel canceled},
			 * the default response handler will be used. Refer to the
			 * [Uploading Dropped or Pasted Files](#!/guide/dev_file_upload) article for more information.
			 *
			 * @since 4.5
			 * @event fileUploadResponse
			 * @member CKEDITOR.editor
			 * @param data All data will be passed to {@link CKEDITOR.fileTools.fileLoader#responseData}.
			 * @param {CKEDITOR.fileTools.fileLoader} data.fileLoader A file loader instance.
			 * @param {String} data.message The message from the server. Needs to be set in the listener &mdash; see the example above.
			 * @param {String} data.fileName The file name on server. Needs to be set in the listener &mdash; see the example above.
			 * @param {String} data.url The URL to the uploaded file. Needs to be set in the listener &mdash; see the example above.
			 */
			editor.on( 'fileUploadResponse', function( evt ) {
				var fileLoader = evt.data.fileLoader,
					xhr = fileLoader.xhr,
					data = evt.data;

				try {
					var response = JSON.parse( xhr.responseText );

					// Error message does not need to mean that upload finished unsuccessfully.
					// It could mean that ex. file name was changes during upload due to naming collision.
					if ( response.error && response.error.message ) {
						data.message = response.error.message;
					}

					// But !uploaded means error.
					if ( !response.uploaded ) {
						evt.cancel();
					} else {
						for ( var i in response ) {
							data[ i ] = response[ i ];
						}
					}
				} catch ( err ) {
					// Response parsing error.
					data.message = fileLoader.lang.filetools.responseError;
					CKEDITOR.warn( 'filetools-response-error', { responseText: xhr.responseText } );

					evt.cancel();
				}
			}, null, null, 999 );
		}
	} );

	/**
	 * File loader repository. It allows you to create and get {@link CKEDITOR.fileTools.fileLoader file loaders}.
	 *
	 * An instance of the repository is available as the {@link CKEDITOR.editor#uploadRepository}.
	 *
	 *		var loader = editor.uploadRepository.create( file );
	 *		loader.loadAndUpload( 'http://foo/bar' );
	 *
	 * To find more information about handling files see the {@link CKEDITOR.fileTools.fileLoader} class.
	 *
	 * @since 4.5
	 * @class CKEDITOR.fileTools.uploadRepository
	 * @mixins CKEDITOR.event
	 * @constructor Creates an instance of the repository.
	 * @param {CKEDITOR.editor} editor Editor instance. Used only to get the language data.
	 */
	function UploadRepository( editor ) {
		this.editor = editor;

		this.loaders = [];
	}

	UploadRepository.prototype = {
		/**
		 * Creates a {@link CKEDITOR.fileTools.fileLoader file loader} instance with a unique ID.
		 * The instance can be later retrieved from the repository using the {@link #loaders} array.
		 *
		 * Fires the {@link CKEDITOR.fileTools.uploadRepository#instanceCreated instanceCreated} event.
		 *
		 * @param {Blob/String} fileOrData See {@link CKEDITOR.fileTools.fileLoader}.
		 * @param {String} fileName See {@link CKEDITOR.fileTools.fileLoader}.
		 * @returns {CKEDITOR.fileTools.fileLoader} The created file loader instance.
		 */
		create: function( fileOrData, fileName ) {
			var id = this.loaders.length,
				loader = new FileLoader( this.editor, fileOrData, fileName );

			loader.id = id;
			this.loaders[ id ] = loader;

			this.fire( 'instanceCreated', loader );

			return loader;
		},

		/**
		 * Returns `true` if all loaders finished their jobs.
		 *
		 * @returns {Boolean} `true` if all loaders finished their job, `false` otherwise.
		 */
		isFinished: function() {
			for ( var id = 0; id < this.loaders.length; ++id ) {
				if ( !this.loaders[ id ].isFinished() ) {
					return false;
				}
			}

			return true;
		}

		/**
		 * Array of loaders created by the {@link #create} method. Loaders' {@link CKEDITOR.fileTools.fileLoader#id IDs}
		 * are indexes.
		 *
		 * @readonly
		 * @property {CKEDITOR.fileTools.fileLoader[]} loaders
		 */

		/**
		 * Event fired when the {@link CKEDITOR.fileTools.fileLoader file loader} is created.
		 *
		 * @event instanceCreated
		 * @param {CKEDITOR.fileTools.fileLoader} data Created file loader.
		 */
	};

	/**
	 * The `FileLoader` class is a wrapper which handles two file operations: loading the content of the file stored on
	 * the user's device into the memory and uploading the file to the server.
	 *
	 * There are two possible ways to crate a `FileLoader` instance: with a [Blob](https://developer.mozilla.org/en/docs/Web/API/Blob)
	 * (e.g. acquired from the {@link CKEDITOR.plugins.clipboard.dataTransfer#getFile} method) or with data as a Base64 string.
	 * Note that if the constructor gets the data as a Base64 string, there is no need to load the data, the data is already loaded.
	 *
	 * The `FileLoader` is created for a single load and upload process so if you abort the process,
	 * you need to create a new `FileLoader`.
	 *
	 * All process parameters are stored in public properties.
	 *
	 * `FileLoader` implements events so you can listen to them to react to changes. There are two types of events:
	 * events to notify the listeners about changes and an event that lets the listeners synchronize with current {@link #status}.
	 *
	 * The first group of events contains {@link #event-loading}, {@link #event-loaded}, {@link #event-uploading},
	 * {@link #event-uploaded}, {@link #event-error} and {@link #event-abort}. These events are called only once,
	 * when the {@link #status} changes.
	 *
	 * The second type is the {@link #event-update} event. It is fired every time the {@link #status} changes, the progress changes
	 * or the {@link #method-update} method is called. Is is created to synchronize the visual representation of the loader with
	 * its status. For example if the dialog window shows the upload progress, it should be refreshed on
	 * the {@link #event-update} listener. Then when the user closes and reopens this dialog, the {@link #method-update} method should
	 * be called to refresh the progress.
	 *
	 * Default request and response formats will work with CKFinder 2.4.3 and above. If you need a custom request
	 * or response handling you need to overwrite the default behavior using the {@link CKEDITOR.editor#fileUploadRequest} and
	 * {@link CKEDITOR.editor#fileUploadResponse} events. For more information see their documentation.
	 *
	 * To create a `FileLoader` instance, use the {@link CKEDITOR.fileTools.uploadRepository} class.
	 *
	 * Here is a simple `FileLoader` usage example:
	 *
	 *		editor.on( 'paste', function( evt ) {
	 *			for ( var i = 0; i < evt.data.dataTransfer.getFilesCount(); i++ ) {
	 *				var file = evt.data.dataTransfer.getFile( i );
	 *
	 *				if ( CKEDITOR.fileTools.isTypeSupported( file, /image\/png/ ) ) {
	 *					var loader = editor.uploadRepository.create( file );
	 *
	 *					loader.on( 'update', function() {
	 *						document.getElementById( 'uploadProgress' ).innerHTML = loader.status;
	 *					} );
	 *
	 *					loader.on( 'error', function() {
	 *						alert( 'Error!' );
	 *					} );
	 *
	 *					loader.loadAndUpload( 'http://upload.url/' );
	 *
	 *					evt.data.dataValue += 'loading...'
	 *				}
	 *			}
	 *		} );
	 *
	 * Note that `FileLoader` uses the native file API which is supported **since Internet Explorer 10**.
	 *
	 * @since 4.5
	 * @class CKEDITOR.fileTools.fileLoader
	 * @mixins CKEDITOR.event
	 * @constructor Creates an instance of the class and sets initial values for all properties.
	 * @param {CKEDITOR.editor} editor The editor instance. Used only to get language data.
	 * @param {Blob/String} fileOrData A [blob object](https://developer.mozilla.org/en/docs/Web/API/Blob) or a data
	 * string encoded with Base64.
	 * @param {String} [fileName] The file name. If not set and the second parameter is a file, then its name will be used.
	 * If not set and the second parameter is a Base64 data string, then the file name will be created based on
	 * the {@link CKEDITOR.config#fileTools_defaultFileName} option.
	 */
	function FileLoader( editor, fileOrData, fileName ) {
		var mimeParts,
			defaultFileName = editor.config.fileTools_defaultFileName;

		this.editor = editor;
		this.lang = editor.lang;

		if ( typeof fileOrData === 'string' ) {
			// Data is already loaded from disc.
			this.data = fileOrData;
			this.file = dataToFile( this.data );
			this.total = this.file.size;
			this.loaded = this.total;
		} else {
			this.data = null;
			this.file = fileOrData;
			this.total = this.file.size;
			this.loaded = 0;
		}

		if ( fileName ) {
			this.fileName = fileName;
		} else if ( this.file.name ) {
			this.fileName = this.file.name;
		} else {
			mimeParts = this.file.type.split( '/' );

			if ( defaultFileName ) {
				mimeParts[ 0 ] = defaultFileName;
			}

			this.fileName = mimeParts.join( '.' );
		}

		this.uploaded = 0;
		this.uploadTotal = null;

		this.responseData = null;

		this.status = 'created';

		this.abort = function() {
			this.changeStatus( 'abort' );
		};
	}

	/**
	 * The loader status. Possible values:
	 *
	 * * `created` &ndash; The loader was created, but neither load nor upload started.
	 * * `loading` &ndash; The file is being loaded from the user's storage.
	 * * `loaded` &ndash; The file was loaded, the process is finished.
	 * * `uploading` &ndash; The file is being uploaded to the server.
	 * * `uploaded` &ndash; The file was uploaded, the process is finished.
	 * * `error` &ndash; The process stops because of an error, more details are available in the {@link #message} property.
	 * * `abort` &ndash; The process was stopped by the user.
	 *
	 * @property {String} status
	 */

	/**
	 * String data encoded with Base64. If the `FileLoader` is created with a Base64 string, the `data` is that string.
	 * If a file was passed to the constructor, the data is `null` until loading is completed.
	 *
	 * @readonly
	 * @property {String} data
	 */

	/**
	 * File object which represents the handled file. This property is set for both constructor options (file or data).
	 *
	 * @readonly
	 * @property {Blob} file
	 */

	/**
	 * The name of the file. If there is no file name, it is created by using the
	 * {@link CKEDITOR.config#fileTools_defaultFileName} option.
	 *
	 * @readonly
	 * @property {String} fileName
	 */

	/**
	 * The number of loaded bytes. If the `FileLoader` was created with a data string,
	 * the loaded value equals the {@link #total} value.
	 *
	 * @readonly
	 * @property {Number} loaded
	 */

	/**
	 * The number of uploaded bytes.
	 *
	 * @readonly
	 * @property {Number} uploaded
	 */

	/**
	 * The total file size in bytes.
	 *
	 * @readonly
	 * @property {Number} total
	 */

	/**
	 * All data received in the response from the server. If the server returns additional data, it will be available
	 * in this property.
	 *
	 * It contains all data set in the {@link CKEDITOR.editor#fileUploadResponse} event listener.
	 *
	 * @readonly
	 * @property {Object} responseData
	 */

	/**
	 * The total size of upload data in bytes.
	 * If the `xhr.upload` object is present, this value will indicate the total size of the request payload, not only the file
	 * size itself. If the `xhr.upload` object is not available and the real upload size cannot be obtained, this value will
	 * be equal to {@link #total}. It has a `null` value until the upload size is known.
	 *
	 * 		loader.on( 'update', function() {
	 * 			// Wait till uploadTotal is present.
	 * 			if ( loader.uploadTotal ) {
	 * 				console.log( 'uploadTotal: ' + loader.uploadTotal );
	 * 			}
	 * 		});
	 *
	 * @readonly
	 * @property {Number} uploadTotal
	 */

	/**
	 * The error message or additional information received from the server.
	 *
	 * @readonly
	 * @property {String} message
	 */

	/**
	 * The URL to the file when it is uploaded or received from the server.
	 *
	 * @readonly
	 * @property {String} url
	 */

	/**
	 * The target of the upload.
	 *
	 * @readonly
	 * @property {String} uploadUrl
	 */

	/**
	 *
	 * Native `FileReader` reference used to load the file.
	 *
	 * @readonly
	 * @property {FileReader} reader
	 */

	/**
	 * Native `XMLHttpRequest` reference used to upload the file.
	 *
	 * @readonly
	 * @property {XMLHttpRequest} xhr
	 */

	/**
	 * If `FileLoader` was created using {@link CKEDITOR.fileTools.uploadRepository},
	 * it gets an identifier which is stored in this property.
	 *
	 * @readonly
	 * @property {Number} id
	 */

	/**
	 * Aborts the process.
	 *
	 * This method has a different behavior depending on the current {@link #status}.
	 *
	 * * If the {@link #status} is `loading` or `uploading`, current operation will be aborted.
	 * * If the {@link #status} is `created`, `loading` or `uploading`, the {@link #status} will be changed to `abort`
	 * and the {@link #event-abort} event will be called.
	 * * If the {@link #status} is `loaded`, `uploaded`, `error` or `abort`, this method will do nothing.
	 *
	 * @method abort
	 */

	FileLoader.prototype = {
		/**
		 * Loads a file from the storage on the user's device to the `data` attribute and uploads it to the server.
		 *
		 * The order of {@link #status statuses} for a successful load and upload is:
		 *
		 * * `created`,
		 * * `loading`,
		 * * `uploading`,
		 * * `uploaded`.
		 *
		 * @param {String} url The upload URL.
		 * @param {Object} [additionalRequestParameters] Additional parameters that would be passed to
	 	 * the {@link CKEDITOR.editor#fileUploadRequest} event.
		 */
		loadAndUpload: function( url, additionalRequestParameters ) {
			var loader = this;

			this.once( 'loaded', function( evt ) {
				// Cancel both 'loaded' and 'update' events,
				// because 'loaded' is terminated state.
				evt.cancel();

				loader.once( 'update', function( evt ) {
					evt.cancel();
				}, null, null, 0 );

				// Start uploading.
				loader.upload( url, additionalRequestParameters );
			}, null, null, 0 );

			this.load();
		},

		/**
		 * Loads a file from the storage on the user's device to the `data` attribute.
		 *
		 * The order of the {@link #status statuses} for a successful load is:
		 *
		 * * `created`,
		 * * `loading`,
		 * * `loaded`.
		 */
		load: function() {
			var loader = this;

			this.reader = new FileReader();

			var reader = this.reader;

			loader.changeStatus( 'loading' );

			this.abort = function() {
				loader.reader.abort();
			};

			reader.onabort = function() {
				loader.changeStatus( 'abort' );
			};

			reader.onerror = function() {
				loader.message = loader.lang.filetools.loadError;
				loader.changeStatus( 'error' );
			};

			reader.onprogress = function( evt ) {
				loader.loaded = evt.loaded;
				loader.update();
			};

			reader.onload = function() {
				loader.loaded = loader.total;
				loader.data = reader.result;
				loader.changeStatus( 'loaded' );
			};

			reader.readAsDataURL( this.file );
		},

		/**
		 * Uploads a file to the server.
		 *
		 * The order of the {@link #status statuses} for a successful upload is:
		 *
		 * * `created`,
		 * * `uploading`,
		 * * `uploaded`.
		 *
		 * @param {String} url The upload URL.
		 * @param {Object} [additionalRequestParameters] Additional data that would be passed to
	 	 * the {@link CKEDITOR.editor#fileUploadRequest} event.
		 */
		upload: function( url, additionalRequestParameters ) {
			var requestData = additionalRequestParameters || {};

			if ( !url ) {
				this.message = this.lang.filetools.noUrlError;
				this.changeStatus( 'error' );
			} else {
				this.uploadUrl = url;

				this.xhr = new XMLHttpRequest();
				this.attachRequestListeners();

				if ( this.editor.fire( 'fileUploadRequest', { fileLoader: this, requestData: requestData } ) ) {
					this.changeStatus( 'uploading' );
				}
			}
		},

		/**
		 * Attaches listeners to the XML HTTP request object.
		 *
		 * @private
		 * @param {XMLHttpRequest} xhr XML HTTP request object.
		 */
		attachRequestListeners: function() {
			var loader = this,
				xhr = this.xhr;

			loader.abort = function() {
				xhr.abort();
				onAbort();
			};

			xhr.onerror = onError;
			xhr.onabort = onAbort;

			// http://dev.ckeditor.com/ticket/13533 - When xhr.upload is present attach onprogress, onerror and onabort functions to get actual upload
			// information.
			if ( xhr.upload ) {
				xhr.upload.onprogress = function( evt ) {
					if ( evt.lengthComputable ) {
						// Set uploadTotal with correct data.
						if ( !loader.uploadTotal ) {
							loader.uploadTotal = evt.total;
						}
						loader.uploaded = evt.loaded;
						loader.update();
					}
				};

				xhr.upload.onerror = onError;
				xhr.upload.onabort = onAbort;

			} else {
				// http://dev.ckeditor.com/ticket/13533 - If xhr.upload is not supported - fire update event anyway and set uploadTotal to file size.
				loader.uploadTotal = loader.total;
				loader.update();
			}

			xhr.onload = function() {
				// http://dev.ckeditor.com/ticket/13433 - Call update at the end of the upload. When xhr.upload object is not supported there will be
				// no update events fired during the whole process.
				loader.update();

				// http://dev.ckeditor.com/ticket/13433 - Check if loader was not aborted during last update.
				if ( loader.status == 'abort' ) {
					return;
				}

				loader.uploaded = loader.uploadTotal;

				if ( xhr.status < 200 || xhr.status > 299 ) {
					loader.message = loader.lang.filetools[ 'httpError' + xhr.status ];
					if ( !loader.message ) {
						loader.message = loader.lang.filetools.httpError.replace( '%1', xhr.status );
					}
					loader.changeStatus( 'error' );
				} else {
					var data = {
							fileLoader: loader
						},
						// Values to copy from event to FileLoader.
						valuesToCopy = [ 'message', 'fileName', 'url' ],
						success = loader.editor.fire( 'fileUploadResponse', data );

					for ( var i = 0; i < valuesToCopy.length; i++ ) {
						var key = valuesToCopy[ i ];
						if ( typeof data[ key ] === 'string' ) {
							loader[ key ] = data[ key ];
						}
					}

					// The whole response is also hold for use by uploadwidgets (http://dev.ckeditor.com/ticket/13519).
					loader.responseData = data;
					// But without reference to the loader itself.
					delete loader.responseData.fileLoader;

					if ( success === false ) {
						loader.changeStatus( 'error' );
					} else {
						loader.changeStatus( 'uploaded' );
					}
				}
			};

			function onError() {
				// Prevent changing status twice, when HHR.error and XHR.upload.onerror could be called together.
				if ( loader.status == 'error' ) {
					return;
				}

				loader.message = loader.lang.filetools.networkError;
				loader.changeStatus( 'error' );
			}

			function onAbort() {
				// Prevent changing status twice, when HHR.onabort and XHR.upload.onabort could be called together.
				if ( loader.status == 'abort' ) {
					return;
				}
				loader.changeStatus( 'abort' );
			}
		},

		/**
		 * Changes {@link #status} to the new status, updates the {@link #method-abort} method if needed and fires two events:
		 * new status and {@link #event-update}.
		 *
		 * @private
		 * @param {String} newStatus New status to be set.
		 */
		changeStatus: function( newStatus ) {
			this.status = newStatus;

			if ( newStatus == 'error' || newStatus == 'abort' ||
				newStatus == 'loaded' || newStatus == 'uploaded' ) {
				this.abort = function() {};
			}

			this.fire( newStatus );
			this.update();
		},

		/**
		 * Updates the state of the `FileLoader` listeners. This method should be called if the state of the visual representation
		 * of the upload process is out of synchronization and needs to be refreshed (e.g. because of an undo operation or
		 * because the dialog window with the upload is closed and reopened). Fires the {@link #event-update} event.
		 */
		update: function() {
			this.fire( 'update' );
		},

		/**
		 * Returns `true` if the loading and uploading finished (successfully or not), so the {@link #status} is
		 * `loaded`, `uploaded`, `error` or `abort`.
		 *
		 * @returns {Boolean} `true` if the loading and uploading finished.
		 */
		isFinished: function() {
			return !!this.status.match( /^(?:loaded|uploaded|error|abort)$/ );
		}

		/**
		 * Event fired when the {@link #status} changes to `loading`. It will be fired once for the `FileLoader`.
		 *
		 * @event loading
		 */

		/**
		 * Event fired when the {@link #status} changes to `loaded`. It will be fired once for the `FileLoader`.
		 *
		 * @event loaded
		 */

		/**
		 * Event fired when the {@link #status} changes to `uploading`. It will be fired once for the `FileLoader`.
		 *
		 * @event uploading
		 */

		/**
		 * Event fired when the {@link #status} changes to `uploaded`. It will be fired once for the `FileLoader`.
		 *
		 * @event uploaded
		 */

		/**
		 * Event fired when the {@link #status} changes to `error`. It will be fired once for the `FileLoader`.
		 *
		 * @event error
		 */

		/**
		 * Event fired when the {@link #status} changes to `abort`. It will be fired once for the `FileLoader`.
		 *
		 * @event abort
		 */

		/**
		 * Event fired every time the `FileLoader` {@link #status} or progress changes or the {@link #method-update} method is called.
		 * This event was designed to allow showing the visualization of the progress and refresh that visualization
		 * every time the status changes. Note that multiple `update` events may be fired with the same status.
		 *
		 * @event update
		 */
	};

	CKEDITOR.event.implementOn( UploadRepository.prototype );
	CKEDITOR.event.implementOn( FileLoader.prototype );

	var base64HeaderRegExp = /^data:(\S*?);base64,/;

	// Transforms Base64 string data into file and creates name for that file based on the mime type.
	//
	// @private
	// @param {String} data Base64 string data.
	// @returns {Blob} File.
	function dataToFile( data ) {
		var contentType = data.match( base64HeaderRegExp )[ 1 ],
			base64Data = data.replace( base64HeaderRegExp, '' ),
			byteCharacters = atob( base64Data ),
			byteArrays = [],
			sliceSize = 512,
			offset, slice, byteNumbers, i, byteArray;

		for ( offset = 0; offset < byteCharacters.length; offset += sliceSize ) {
			slice = byteCharacters.slice( offset, offset + sliceSize );

			byteNumbers = new Array( slice.length );
			for ( i = 0; i < slice.length; i++ ) {
				byteNumbers[ i ] = slice.charCodeAt( i );
			}

			byteArray = new Uint8Array( byteNumbers );

			byteArrays.push( byteArray );
		}

		return new Blob( byteArrays, { type: contentType } );
	}

	//
	// PUBLIC API -------------------------------------------------------------
	//

	// Two plugins extend this object.
	if ( !CKEDITOR.fileTools ) {
		/**
		 * Helpers to load and upload a file.
		 *
		 * @since 4.5
		 * @singleton
		 * @class CKEDITOR.fileTools
		 */
		CKEDITOR.fileTools = {};
	}

	CKEDITOR.tools.extend( CKEDITOR.fileTools, {
		uploadRepository: UploadRepository,
		fileLoader: FileLoader,

		/**
		 * Gets the upload URL from the {@link CKEDITOR.config configuration}. Because of backward compatibility
		 * the URL can be set using multiple configuration options.
		 *
		 * If the `type` is defined, then four configuration options will be checked in the following order
		 * (examples for `type='image'`):
		 *
		 * * `[type]UploadUrl`, e.g. {@link CKEDITOR.config#imageUploadUrl},
		 * * {@link CKEDITOR.config#uploadUrl},
		 * * `filebrowser[uppercased type]uploadUrl`, e.g. {@link CKEDITOR.config#filebrowserImageUploadUrl},
		 * * {@link CKEDITOR.config#filebrowserUploadUrl}.
		 *
		 * If the `type` is not defined, two configuration options will be checked:
		 *
		 * * {@link CKEDITOR.config#uploadUrl},
		 * * {@link CKEDITOR.config#filebrowserUploadUrl}.
		 *
		 * `filebrowser[type]uploadUrl` and `filebrowserUploadUrl` are checked for backward compatibility with the
		 * `filebrowser` plugin.
		 *
		 * For both `filebrowser[type]uploadUrl` and `filebrowserUploadUrl` `&responseType=json` is added to the end of the URL.
		 *
		 * @param {Object} config The configuration file.
		 * @param {String} [type] Upload file type.
		 * @returns {String/null} Upload URL or `null` if none of the configuration options were defined.
		 */
		getUploadUrl: function( config, type ) {
			var capitalize = CKEDITOR.tools.capitalize;

			if ( type && config[ type + 'UploadUrl' ] ) {
				return config[ type + 'UploadUrl' ];
			} else if ( config.uploadUrl ) {
				return config.uploadUrl;
			} else if ( type && config[ 'filebrowser' + capitalize( type, 1 ) + 'UploadUrl' ] ) {
				return config[ 'filebrowser' + capitalize( type, 1 ) + 'UploadUrl' ] + '&responseType=json';
			} else if ( config.filebrowserUploadUrl ) {
				return config.filebrowserUploadUrl + '&responseType=json';
			}

			return null;
		},

		/**
		 * Checks if the MIME type of the given file is supported.
		 *
		 * 		CKEDITOR.fileTools.isTypeSupported( { type: 'image/png' }, /image\/(png|jpeg)/ ); // true
		 * 		CKEDITOR.fileTools.isTypeSupported( { type: 'image/png' }, /image\/(gif|jpeg)/ ); // false
		 *
		 * @param {Blob} file The file to check.
		 * @param {RegExp} supportedTypes A regular expression to check the MIME type of the file.
		 * @returns {Boolean} `true` if the file type is supported.
		 */
		isTypeSupported: function( file, supportedTypes ) {
			return !!file.type.match( supportedTypes );
		}
	} );
} )();

/**
 * The URL where files should be uploaded.
 *
 * An empty string means that the option is disabled.
 *
 * @since 4.5
 * @cfg {String} [uploadUrl='']
 * @member CKEDITOR.config
 */

/**
 * Default file name (without extension) that will be used for files created from a Base64 data string
 * (for example for files pasted into the editor).
 * This name will be combined with the MIME type to create the full file name with the extension.
 *
 * If `fileTools_defaultFileName` is set to `default-name` and data's MIME type is `image/png`,
 * the resulting file name will be `default-name.png`.
 *
 * If `fileTools_defaultFileName` is not set, the file name will be created using only its MIME type.
 * For example for `image/png` the file name will be `image.png`.
 *
 * @since 4.5.3
 * @cfg {String} [fileTools_defaultFileName='']
 * @member CKEDITOR.config
 */
