/**
 * @license Copyright (c) 2003-2017, CKSource - Frederico Knabben. All rights reserved.
 * For licensing, see LICENSE.md or http://ckeditor.com/license
 */

/**
 * @ignore
 * File overview: Clipboard support.
 */

//
// COPY & PASTE EXECUTION FLOWS:
// -- CTRL+C
//		* if ( isCustomCopyCutSupported )
//			* dataTransfer.setData( 'text/html', getSelectedHtml )
//		* else
//			* browser's default behavior
// -- CTRL+X
//		* listen onKey (onkeydown)
//		* fire 'saveSnapshot' on editor
//		* if ( isCustomCopyCutSupported )
//			* dataTransfer.setData( 'text/html', getSelectedHtml )
//			* extractSelectedHtml // remove selected contents
//		* else
//			* browser's default behavior
//		* deferred second 'saveSnapshot' event
// -- CTRL+V
//		* listen onKey (onkeydown)
//		* simulate 'beforepaste' for non-IEs on editable
//		* listen 'onpaste' on editable ('onbeforepaste' for IE)
//		* fire 'beforePaste' on editor
//		* if ( !canceled && ( htmlInDataTransfer || !external paste) && dataTransfer is not empty ) getClipboardDataByPastebin
//		* fire 'paste' on editor
//		* !canceled && fire 'afterPaste' on editor
// -- Copy command
//		* tryToCutCopy
//			* execCommand
//		* !success && notification
// -- Cut command
//		* fixCut
//		* tryToCutCopy
//			* execCommand
//		* !success && notification
// -- Paste command
//		* fire 'paste' on editable ('beforepaste' for IE)
//		* !canceled && execCommand 'paste'
// -- Paste from native context menu & menubar
//		(Fx & Webkits are handled in 'paste' default listener.
//		Opera cannot be handled at all because it doesn't fire any events
//		Special treatment is needed for IE, for which is this part of doc)
//		* listen 'onpaste'
//		* cancel native event
//		* fire 'beforePaste' on editor
//		* if ( !canceled && ( htmlInDataTransfer || !external paste) && dataTransfer is not empty ) getClipboardDataByPastebin
//		* execIECommand( 'paste' ) -> this fires another 'paste' event, so cancel it
//		* fire 'paste' on editor
//		* !canceled && fire 'afterPaste' on editor
//
//
// PASTE EVENT - PREPROCESSING:
// -- Possible dataValue types: auto, text, html.
// -- Possible dataValue contents:
//		* text (possible \n\r)
//		* htmlified text (text + br,div,p - no presentational markup & attrs - depends on browser)
//		* html
// -- Possible flags:
//		* htmlified - if true then content is a HTML even if no markup inside. This flag is set
//			for content from editable pastebins, because they 'htmlify' pasted content.
//
// -- Type: auto:
//		* content: htmlified text ->	filter, unify text markup (brs, ps, divs), set type: text
//		* content: html ->				filter, set type: html
// -- Type: text:
//		* content: htmlified text ->	filter, unify text markup
//		* content: html ->				filter, strip presentational markup, unify text markup
// -- Type: html:
//		* content: htmlified text ->	filter, unify text markup
//		* content: html ->				filter
//
// -- Phases:
//		* if dataValue is empty copy data from dataTransfer to dataValue (priority 1)
//		* filtering (priorities 3-5) - e.g. pastefromword filters
//		* content type sniffing (priority 6)
//		* markup transformations for text (priority 6)
//
// DRAG & DROP EXECUTION FLOWS:
// -- Drag
//		* save to the global object:
//			* drag timestamp (with 'cke-' prefix),
//			* selected html,
//			* drag range,
//			* editor instance.
//		* put drag timestamp into event.dataTransfer.text
// -- Drop
//		* if events text == saved timestamp && editor == saved editor
//			internal drag & drop occurred
//			* getRangeAtDropPosition
//			* create bookmarks for drag and drop ranges starting from the end of the document
//			* dragRange.deleteContents()
//			* fire 'paste' with saved html and drop range
//		* if events text == saved timestamp && editor != saved editor
//			cross editor drag & drop occurred
//			* getRangeAtDropPosition
//			* fire 'paste' with saved html
//			* dragRange.deleteContents()
//			* FF: refreshCursor on afterPaste
//		* if events text != saved timestamp
//			drop form external source occurred
//			* getRangeAtDropPosition
//			* if event contains html data then fire 'paste' with html
//			* else if event contains text data then fire 'paste' with encoded text
//			* FF: refreshCursor on afterPaste

'use strict';

( function() {
	// Register the plugin.
	CKEDITOR.plugins.add( 'clipboard', {
		requires: 'notification,toolbar',
		// jscs:disable maximumLineLength
		lang: 'af,ar,az,bg,bn,bs,ca,cs,cy,da,de,de-ch,el,en,en-au,en-ca,en-gb,eo,es,es-mx,et,eu,fa,fi,fo,fr,fr-ca,gl,gu,he,hi,hr,hu,id,is,it,ja,ka,km,ko,ku,lt,lv,mk,mn,ms,nb,nl,no,oc,pl,pt,pt-br,ro,ru,si,sk,sl,sq,sr,sr-latn,sv,th,tr,tt,ug,uk,vi,zh,zh-cn', // %REMOVE_LINE_CORE%
		// jscs:enable maximumLineLength
		icons: 'copy,copy-rtl,cut,cut-rtl,paste,paste-rtl', // %REMOVE_LINE_CORE%
		hidpi: true, // %REMOVE_LINE_CORE%
		init: function( editor ) {
			var filterType,
				filtersFactory = filtersFactoryFactory();

			if ( editor.config.forcePasteAsPlainText ) {
				filterType = 'plain-text';
			} else if ( editor.config.pasteFilter ) {
				filterType = editor.config.pasteFilter;
			}
			// On Webkit the pasteFilter defaults 'semantic-content' because pasted data is so terrible
			// that it must be always filtered.
			else if ( CKEDITOR.env.webkit && !( 'pasteFilter' in editor.config ) ) {
				filterType = 'semantic-content';
			}

			editor.pasteFilter = filtersFactory.get( filterType );

			initPasteClipboard( editor );
			initDragDrop( editor );

			// Convert image file (if present) to base64 string for Firefox. Do it as the first
			// step as the conversion is asynchronous and should hold all further paste processing.
			if ( CKEDITOR.env.gecko ) {
				var supportedImageTypes = [ 'image/png', 'image/jpeg', 'image/gif' ],
					latestId;

				editor.on( 'paste', function( evt ) {
					var dataObj = evt.data,
						data = dataObj.dataValue,
						dataTransfer = dataObj.dataTransfer;

					// If data empty check for image content inside data transfer. http://dev.ckeditor.com/ticket/16705
					if ( !data && dataObj.method == 'paste' && dataTransfer && dataTransfer.getFilesCount() == 1 && latestId != dataTransfer.id ) {
						var file = dataTransfer.getFile( 0 );

						if ( CKEDITOR.tools.indexOf( supportedImageTypes, file.type ) != -1 ) {
							var fileReader = new FileReader();

							// Convert image file to img tag with base64 image.
							fileReader.addEventListener( 'load', function() {
								evt.data.dataValue = '<img src="' + fileReader.result + '" />';
								editor.fire( 'paste', evt.data );
							}, false );

							// Proceed with normal flow if reading file was aborted.
							fileReader.addEventListener( 'abort', function() {
								editor.fire( 'paste', evt.data );
							}, false );

							// Proceed with normal flow if reading file failed.
							fileReader.addEventListener( 'error', function() {
								editor.fire( 'paste', evt.data );
							}, false );

							fileReader.readAsDataURL( file );

							latestId = dataObj.dataTransfer.id;

							evt.stop();
						}
					}
				}, null, null, 1 );
			}

			editor.on( 'paste', function( evt ) {
				// Init `dataTransfer` if `paste` event was fired without it, so it will be always available.
				if ( !evt.data.dataTransfer ) {
					evt.data.dataTransfer = new CKEDITOR.plugins.clipboard.dataTransfer();
				}

				// If dataValue is already set (manually or by paste bin), so do not override it.
				if ( evt.data.dataValue ) {
					return;
				}

				var dataTransfer = evt.data.dataTransfer,
					// IE support only text data and throws exception if we try to get html data.
					// This html data object may also be empty if we drag content of the textarea.
					value = dataTransfer.getData( 'text/html' );

				if ( value ) {
					evt.data.dataValue = value;
					evt.data.type = 'html';
				} else {
					// Try to get text data otherwise.
					value = dataTransfer.getData( 'text/plain' );

					if ( value ) {
						evt.data.dataValue = editor.editable().transformPlainTextToHtml( value );
						evt.data.type = 'text';
					}
				}
			}, null, null, 1 );

			editor.on( 'paste', function( evt ) {
				var data = evt.data.dataValue,
					blockElements = CKEDITOR.dtd.$block;

				// Filter webkit garbage.
				if ( data.indexOf( 'Apple-' ) > -1 ) {
					// Replace special webkit's &nbsp; with simple space, because webkit
					// produces them even for normal spaces.
					data = data.replace( /<span class="Apple-converted-space">&nbsp;<\/span>/gi, ' ' );

					// Strip <span> around white-spaces when not in forced 'html' content type.
					// This spans are created only when pasting plain text into Webkit,
					// but for safety reasons remove them always.
					if ( evt.data.type != 'html' ) {
						data = data.replace( /<span class="Apple-tab-span"[^>]*>([^<]*)<\/span>/gi, function( all, spaces ) {
							// Replace tabs with 4 spaces like Fx does.
							return spaces.replace( /\t/g, '&nbsp;&nbsp; &nbsp;' );
						} );
					}

					// This br is produced only when copying & pasting HTML content.
					if ( data.indexOf( '<br class="Apple-interchange-newline">' ) > -1 ) {
						evt.data.startsWithEOL = 1;
						evt.data.preSniffing = 'html'; // Mark as not text.
						data = data.replace( /<br class="Apple-interchange-newline">/, '' );
					}

					// Remove all other classes.
					data = data.replace( /(<[^>]+) class="Apple-[^"]*"/gi, '$1' );
				}

				// Strip editable that was copied from inside. (http://dev.ckeditor.com/ticket/9534)
				if ( data.match( /^<[^<]+cke_(editable|contents)/i ) ) {
					var tmp,
						editable_wrapper,
						wrapper = new CKEDITOR.dom.element( 'div' );

					wrapper.setHtml( data );
					// Verify for sure and check for nested editor UI parts. (http://dev.ckeditor.com/ticket/9675)
					while ( wrapper.getChildCount() == 1 &&
							( tmp = wrapper.getFirst() ) &&
							tmp.type == CKEDITOR.NODE_ELEMENT &&	// Make sure first-child is element.
							( tmp.hasClass( 'cke_editable' ) || tmp.hasClass( 'cke_contents' ) ) ) {
						wrapper = editable_wrapper = tmp;
					}

					// If editable wrapper was found strip it and bogus <br> (added on FF).
					if ( editable_wrapper )
						data = editable_wrapper.getHtml().replace( /<br>$/i, '' );
				}

				if ( CKEDITOR.env.ie ) {
					// &nbsp; <p> -> <p> (br.cke-pasted-remove will be removed later)
					data = data.replace( /^&nbsp;(?: |\r\n)?<(\w+)/g, function( match, elementName ) {
						if ( elementName.toLowerCase() in blockElements ) {
							evt.data.preSniffing = 'html'; // Mark as not a text.
							return '<' + elementName;
						}
						return match;
					} );
				} else if ( CKEDITOR.env.webkit ) {
					// </p><div><br></div> -> </p><br>
					// We don't mark br, because this situation can happen for htmlified text too.
					data = data.replace( /<\/(\w+)><div><br><\/div>$/, function( match, elementName ) {
						if ( elementName in blockElements ) {
							evt.data.endsWithEOL = 1;
							return '</' + elementName + '>';
						}
						return match;
					} );
				} else if ( CKEDITOR.env.gecko ) {
					// Firefox adds bogus <br> when user pasted text followed by space(s).
					data = data.replace( /(\s)<br>$/, '$1' );
				}

				evt.data.dataValue = data;
			}, null, null, 3 );

			editor.on( 'paste', function( evt ) {
				var dataObj = evt.data,
					type = editor._.nextPasteType || dataObj.type,
					data = dataObj.dataValue,
					trueType,
					// Default is 'html'.
					defaultType = editor.config.clipboard_defaultContentType || 'html',
					transferType = dataObj.dataTransfer.getTransferType( editor );

				// If forced type is 'html' we don't need to know true data type.
				if ( type == 'html' || dataObj.preSniffing == 'html' ) {
					trueType = 'html';
				} else {
					trueType = recogniseContentType( data );
				}

				delete editor._.nextPasteType;

				// Unify text markup.
				if ( trueType == 'htmlifiedtext' ) {
					data = htmlifiedTextHtmlification( editor.config, data );
				}

				// Strip presentational markup & unify text markup.
				// Forced plain text.
				// Note: we do not check dontFilter option in this case, because forcePAPT was implemented
				// before pasteFilter and pasteFilter is automatically used on Webkit&Blink since 4.5, so
				// forcePAPT should have priority as it had before 4.5.
				if ( type == 'text' && trueType == 'html' ) {
					data = filterContent( editor, data, filtersFactory.get( 'plain-text' ) );
				}
				// External paste and pasteFilter exists and filtering isn't disabled.
				else if ( transferType == CKEDITOR.DATA_TRANSFER_EXTERNAL && editor.pasteFilter && !dataObj.dontFilter ) {
					data = filterContent( editor, data, editor.pasteFilter );
				}

				if ( dataObj.startsWithEOL ) {
					data = '<br data-cke-eol="1">' + data;
				}
				if ( dataObj.endsWithEOL ) {
					data += '<br data-cke-eol="1">';
				}

				if ( type == 'auto' ) {
					type = ( trueType == 'html' || defaultType == 'html' ) ? 'html' : 'text';
				}

				dataObj.type = type;
				dataObj.dataValue = data;
				delete dataObj.preSniffing;
				delete dataObj.startsWithEOL;
				delete dataObj.endsWithEOL;
			}, null, null, 6 );

			// Inserts processed data into the editor at the end of the
			// events chain.
			editor.on( 'paste', function( evt ) {
				var data = evt.data;

				if ( data.dataValue ) {
					editor.insertHtml( data.dataValue, data.type, data.range );

					// Defer 'afterPaste' so all other listeners for 'paste' will be fired first.
					// Fire afterPaste only if paste inserted some HTML.
					setTimeout( function() {
						editor.fire( 'afterPaste' );
					}, 0 );
				}
			}, null, null, 1000 );
		}
	} );

	function firePasteEvents( editor, data, withBeforePaste ) {
		if ( !data.type ) {
			data.type = 'auto';
		}

		if ( withBeforePaste ) {
			// Fire 'beforePaste' event so clipboard flavor get customized
			// by other plugins.
			if ( editor.fire( 'beforePaste', data ) === false )
				return false; // Event canceled
		}

		// Do not fire paste if there is no data (dataValue and dataTranfser are empty).
		// This check should be done after firing 'beforePaste' because for native paste
		// 'beforePaste' is by default fired even for empty clipboard.
		if ( !data.dataValue && data.dataTransfer.isEmpty() ) {
			return false;
		}

		if ( !data.dataValue ) {
			data.dataValue = '';
		}

		// Because of FF bug we need to use this hack, otherwise cursor is hidden
		// or it is not possible to move it (http://dev.ckeditor.com/ticket/12420).
		// Also, check that editor.toolbox exists, because the toolbar plugin might not be loaded (http://dev.ckeditor.com/ticket/13305).
		if ( CKEDITOR.env.gecko && data.method == 'drop' && editor.toolbox ) {
			editor.once( 'afterPaste', function() {
				editor.toolbox.focus();
			} );
		}

		return editor.fire( 'paste', data );
	}

	function initPasteClipboard( editor ) {
		var clipboard = CKEDITOR.plugins.clipboard,
			preventBeforePasteEvent = 0,
			preventPasteEvent = 0,
			inReadOnly = 0;

		addListeners();
		addButtonsCommands();

		/**
		 * Gets clipboard data by directly accessing the clipboard (IE only).
		 *
		 *		editor.getClipboardData( function( data ) {
		 *			if ( data )
		 *				alert( data.type + ' ' + data.dataValue );
		 *		} );
		 *
		 * @member CKEDITOR.editor
		 * @param {Function/Object} callbackOrOptions For function, see the `callback` parameter documentation. The object was used before 4.7.0 with the `title` property, to set the paste dialog's title.
		 * @param {Function} callback A function that will be executed with the `data` property of the
		 * {@link CKEDITOR.editor#event-paste paste event} or `null` if none of the capturing methods succeeded.
		 * Since 4.7.0 the `callback` should be provided as a first argument, just like in the example above. This parameter will be removed in
		 * an upcoming major release.
		 */
		editor.getClipboardData = function( callbackOrOptions, callback ) {
			// Options are optional - args shift.
			if ( !callback ) {
				callback = callbackOrOptions;
				callbackOrOptions = null;
			}

			// Listen with maximum priority to handle content before everyone else.
			// This callback will handle paste event that will be fired if direct
			// access to the clipboard succeed in IE.
			editor.on( 'paste', onPaste, null, null, 0 );

			// If command didn't succeed (only IE allows to access clipboard and only if
			// user agrees) invoke callback with null, meaning that paste is not blocked.
			if ( getClipboardDataDirectly() === false ) {
				// Direct access to the clipboard wasn't successful so remove listener.
				editor.removeListener( 'paste', onPaste );

				callback( null );
			}

			function onPaste( evt ) {
				evt.removeListener();
				evt.cancel();
				callback( evt.data );
			}
		};

		function addButtonsCommands() {
			addButtonCommand( 'Cut', 'cut', createCutCopyCmd( 'cut' ), 10, 1 );
			addButtonCommand( 'Copy', 'copy', createCutCopyCmd( 'copy' ), 20, 4 );
			addButtonCommand( 'Paste', 'paste', createPasteCmd(), 30, 8 );

			function addButtonCommand( buttonName, commandName, command, toolbarOrder, ctxMenuOrder ) {
				var lang = editor.lang.clipboard[ commandName ];

				editor.addCommand( commandName, command );
				editor.ui.addButton && editor.ui.addButton( buttonName, {
					label: lang,
					command: commandName,
					toolbar: 'clipboard,' + toolbarOrder
				} );

				// If the "menu" plugin is loaded, register the menu item.
				if ( editor.addMenuItems ) {
					editor.addMenuItem( commandName, {
						label: lang,
						command: commandName,
						group: 'clipboard',
						order: ctxMenuOrder
					} );
				}
			}
		}

		function addListeners() {
			editor.on( 'key', onKey );
			editor.on( 'contentDom', addPasteListenersToEditable );

			// For improved performance, we're checking the readOnly state on selectionChange instead of hooking a key event for that.
			editor.on( 'selectionChange', function( evt ) {
				inReadOnly = evt.data.selection.getRanges()[ 0 ].checkReadOnly();
				setToolbarStates();
			} );

			// If the "contextmenu" plugin is loaded, register the listeners.
			if ( editor.contextMenu ) {
				editor.contextMenu.addListener( function( element, selection ) {
					inReadOnly = selection.getRanges()[ 0 ].checkReadOnly();
					return {
						cut: stateFromNamedCommand( 'cut' ),
						copy: stateFromNamedCommand( 'copy' ),
						paste: stateFromNamedCommand( 'paste' )
					};
				} );
			}
		}

		// Add events listeners to editable.
		function addPasteListenersToEditable() {
			var editable = editor.editable();

			if ( CKEDITOR.plugins.clipboard.isCustomCopyCutSupported ) {
				var initOnCopyCut = function( evt ) {
					// If user tries to cut in read-only editor, we must prevent default action. (http://dev.ckeditor.com/ticket/13872)
					if ( !editor.readOnly || evt.name != 'cut' ) {
						clipboard.initPasteDataTransfer( evt, editor );
					}
					evt.data.preventDefault();
				};

				editable.on( 'copy', initOnCopyCut );
				editable.on( 'cut', initOnCopyCut );

				// Delete content with the low priority so one can overwrite cut data.
				editable.on( 'cut', function() {
					// If user tries to cut in read-only editor, we must prevent default action. (http://dev.ckeditor.com/ticket/13872)
					if ( !editor.readOnly ) {
						editor.extractSelectedHtml();
					}
				}, null, null, 999 );
			}

			// We'll be catching all pasted content in one line, regardless of whether
			// it's introduced by a document command execution (e.g. toolbar buttons) or
			// user paste behaviors (e.g. CTRL+V).
			editable.on( clipboard.mainPasteEvent, function( evt ) {
				if ( clipboard.mainPasteEvent == 'beforepaste' && preventBeforePasteEvent ) {
					return;
				}

				// If you've just asked yourself why preventPasteEventNow() is not here, but
				// in listener for CTRL+V and exec method of 'paste' command
				// you've asked the same question we did.
				//
				// THE ANSWER:
				//
				// First thing to notice - this answer makes sense only for IE,
				// because other browsers don't listen for 'paste' event.
				//
				// What would happen if we move preventPasteEventNow() here?
				// For:
				// * CTRL+V - IE fires 'beforepaste', so we prevent 'paste' and pasteDataFromClipboard(). OK.
				// * editor.execCommand( 'paste' ) - we fire 'beforepaste', so we prevent
				//		'paste' and pasteDataFromClipboard() and doc.execCommand( 'Paste' ). OK.
				// * native context menu - IE fires 'beforepaste', so we prevent 'paste', but unfortunately
				//		on IE we fail with pasteDataFromClipboard() here, because of... we don't know why, but
				//		we just fail, so... we paste nothing. FAIL.
				// * native menu bar - the same as for native context menu.
				//
				// But don't you know any way to distinguish first two cases from last two?
				// Only one - special flag set in CTRL+V handler and exec method of 'paste'
				// command. And that's what we did using preventPasteEventNow().

				pasteDataFromClipboard( evt );
			} );

			// It's not possible to clearly handle all four paste methods (ctrl+v, native menu bar
			// native context menu, editor's command) in one 'paste/beforepaste' event in IE.
			//
			// For ctrl+v & editor's command it's easy to handle pasting in 'beforepaste' listener,
			// so we do this. For another two methods it's better to use 'paste' event.
			//
			// 'paste' is always being fired after 'beforepaste' (except of weird one on opening native
			// context menu), so for two methods handled in 'beforepaste' we're canceling 'paste'
			// using preventPasteEvent state.
			//
			// 'paste' event in IE is being fired before getClipboardDataByPastebin executes its callback.
			//
			// QUESTION: Why didn't you handle all 4 paste methods in handler for 'paste'?
			//		Wouldn't this just be simpler?
			// ANSWER: Then we would have to evt.data.preventDefault() only for native
			//		context menu and menu bar pastes. The same with execIECommand().
			//		That would force us to mark CTRL+V and editor's paste command with
			//		special flag, other than preventPasteEvent. But we still would have to
			//		have preventPasteEvent for the second event fired by execIECommand.
			//		Code would be longer and not cleaner.
			if ( clipboard.mainPasteEvent == 'beforepaste' ) {
				editable.on( 'paste', function( evt ) {
					if ( preventPasteEvent ) {
						return;
					}

					// Cancel next 'paste' event fired by execIECommand( 'paste' )
					// at the end of this callback.
					preventPasteEventNow();

					// Prevent native paste.
					evt.data.preventDefault();

					pasteDataFromClipboard( evt );

					// Force IE to paste content into pastebin so pasteDataFromClipboard will work.
					execIECommand( 'paste' );
				} );

				// If mainPasteEvent is 'beforePaste' (IE before Edge),
				// dismiss the (wrong) 'beforepaste' event fired on context/toolbar menu open. (http://dev.ckeditor.com/ticket/7953)
				editable.on( 'contextmenu', preventBeforePasteEventNow, null, null, 0 );

				editable.on( 'beforepaste', function( evt ) {
					// Do not prevent event on CTRL+V and SHIFT+INS because it blocks paste (http://dev.ckeditor.com/ticket/11970).
					if ( evt.data && !evt.data.$.ctrlKey && !evt.data.$.shiftKey )
						preventBeforePasteEventNow();
				}, null, null, 0 );
			}

			editable.on( 'beforecut', function() {
				!preventBeforePasteEvent && fixCut( editor );
			} );

			var mouseupTimeout;

			// Use editor.document instead of editable in non-IEs for observing mouseup
			// since editable won't fire the event if selection process started within
			// iframe and ended out of the editor (http://dev.ckeditor.com/ticket/9851).
			editable.attachListener( CKEDITOR.env.ie ? editable : editor.document.getDocumentElement(), 'mouseup', function() {
				mouseupTimeout = setTimeout( function() {
					setToolbarStates();
				}, 0 );
			} );

			// Make sure that deferred mouseup callback isn't executed after editor instance
			// had been destroyed. This may happen when editor.destroy() is called in parallel
			// with mouseup event (i.e. a button with onclick callback) (http://dev.ckeditor.com/ticket/10219).
			editor.on( 'destroy', function() {
				clearTimeout( mouseupTimeout );
			} );

			editable.on( 'keyup', setToolbarStates );
		}

		// Create object representing Cut or Copy commands.
		function createCutCopyCmd( type ) {
			return {
				type: type,
				canUndo: type == 'cut', // We can't undo copy to clipboard.
				startDisabled: true,
				fakeKeystroke: type == 'cut' ? CKEDITOR.CTRL + 88 /*X*/ :  CKEDITOR.CTRL + 67 /*C*/,
				exec: function() {
					// Attempts to execute the Cut and Copy operations.
					function tryToCutCopy( type ) {
						if ( CKEDITOR.env.ie )
							return execIECommand( type );

						// non-IEs part
						try {
							// Other browsers throw an error if the command is disabled.
							return editor.document.$.execCommand( type, false, null );
						} catch ( e ) {
							return false;
						}
					}

					this.type == 'cut' && fixCut();

					var success = tryToCutCopy( this.type );

					if ( !success ) {
						// Show cutError or copyError.
						editor.showNotification( editor.lang.clipboard[ this.type + 'Error' ] ); // jshint ignore:line
					}

					return success;
				}
			};
		}

		function createPasteCmd() {
			return {
				// Snapshots are done manually by editable.insertXXX methods.
				canUndo: false,
				async: true,
				fakeKeystroke: CKEDITOR.CTRL + 86 /*V*/,

				/**
				 * The default implementation of the paste command.
				 *
				 * @private
				 * @param {CKEDITOR.editor} editor An instance of the editor where the command is being executed.
				 * @param {Object/String} data If `data` is a string, then it is considered content that is being pasted.
				 * Otherwise it is treated as an object with options.
				 * @param {Boolean/String} [data.notification=true] Content for a notification shown after an unsuccessful
				 * paste attempt. If `false`, the notification will not be displayed. This parameter was added in 4.7.0.
				 * @param {String} [data.type='html'] The type of pasted content. There are two allowed values:
				 * * 'html'
				 * * 'text'
				 * @param {String/Object} data.dataValue Content being pasted. If this parameter is an object, it
				 * is supposed to be a `data` property of the {@link CKEDITOR.editor#paste} event.
				 * @param {CKEDITOR.plugins.clipboard.dataTransfer} data.dataTransfer Data transfer instance connected
				 * with the current paste action.
				 * @member CKEDITOR.editor.commands.paste
				 */
				exec: function( editor, data ) {
					data = typeof data !== 'undefined' && data !== null ? data : {};

					var cmd = this,
						notification = typeof data.notification !== 'undefined' ? data.notification : true,
						forcedType = data.type,
						keystroke = CKEDITOR.tools.keystrokeToString( editor.lang.common.keyboard,
							editor.getCommandKeystroke( this ) ),
						msg = typeof notification === 'string' ? notification : editor.lang.clipboard.pasteNotification
							.replace( /%1/, '<kbd aria-label="' + keystroke.aria + '">' + keystroke.display + '</kbd>' ),
						pastedContent = typeof data === 'string' ? data : data.dataValue;

					function callback( data, withBeforePaste ) {
						withBeforePaste = typeof withBeforePaste !== 'undefined' ? withBeforePaste : true;

						if ( data ) {
							data.method = 'paste';

							if ( !data.dataTransfer ) {
								data.dataTransfer = clipboard.initPasteDataTransfer();
							}

							firePasteEvents( editor, data, withBeforePaste );
						} else if ( notification ) {
							editor.showNotification( msg, 'info', editor.config.clipboard_notificationDuration );
						}

						editor.fire( 'afterCommandExec', {
							name: 'paste',
							command: cmd,
							returnValue: !!data
						} );
					}

					// Force type for the next paste.
					if ( forcedType ) {
						editor._.nextPasteType = forcedType;
					} else {
						delete editor._.nextPasteType;
					}

					if ( typeof pastedContent === 'string' ) {
						callback( {
							dataValue: pastedContent
						} );
					} else {
						editor.getClipboardData( callback );
					}
				}
			};
		}

		function preventPasteEventNow() {
			preventPasteEvent = 1;
			// For safety reason we should wait longer than 0/1ms.
			// We don't know how long execution of quite complex getClipboardData will take
			// and in for example 'paste' listener execCommand() (which fires 'paste') is called
			// after getClipboardData finishes.
			// Luckily, it's impossible to immediately fire another 'paste' event we want to handle,
			// because we only handle there native context menu and menu bar.
			setTimeout( function() {
				preventPasteEvent = 0;
			}, 100 );
		}

		function preventBeforePasteEventNow() {
			preventBeforePasteEvent = 1;
			setTimeout( function() {
				preventBeforePasteEvent = 0;
			}, 10 );
		}

		// Tries to execute any of the paste, cut or copy commands in IE. Returns a
		// boolean indicating that the operation succeeded.
		// @param {String} command *LOWER CASED* name of command ('paste', 'cut', 'copy').
		function execIECommand( command ) {
			var doc = editor.document,
				body = doc.getBody(),
				enabled = false,
				onExec = function() {
					enabled = true;
				};

			// The following seems to be the only reliable way to detect that
			// clipboard commands are enabled in IE. It will fire the
			// onpaste/oncut/oncopy events only if the security settings allowed
			// the command to execute.
			body.on( command, onExec );

			// IE7: document.execCommand has problem to paste into positioned element.
			if ( CKEDITOR.env.version > 7 ) {
				doc.$.execCommand( command );
			} else {
				doc.$.selection.createRange().execCommand( command );
			}

			body.removeListener( command, onExec );

			return enabled;
		}

		// Cutting off control type element in IE standards breaks the selection entirely. (http://dev.ckeditor.com/ticket/4881)
		function fixCut() {
			if ( !CKEDITOR.env.ie || CKEDITOR.env.quirks )
				return;

			var sel = editor.getSelection(),
				control, range, dummy;

			if ( ( sel.getType() == CKEDITOR.SELECTION_ELEMENT ) && ( control = sel.getSelectedElement() ) ) {
				range = sel.getRanges()[ 0 ];
				dummy = editor.document.createText( '' );
				dummy.insertBefore( control );
				range.setStartBefore( dummy );
				range.setEndAfter( control );
				sel.selectRanges( [ range ] );

				// Clear up the fix if the paste wasn't succeeded.
				setTimeout( function() {
					// Element still online?
					if ( control.getParent() ) {
						dummy.remove();
						sel.selectElement( control );
					}
				}, 0 );
			}
		}

		// Allow to peek clipboard content by redirecting the
		// pasting content into a temporary bin and grab the content of it.
		function getClipboardDataByPastebin( evt, callback ) {
			var doc = editor.document,
				editable = editor.editable(),
				cancel = function( evt ) {
					evt.cancel();
				},
				blurListener;

			// Avoid recursions on 'paste' event or consequent paste too fast. (http://dev.ckeditor.com/ticket/5730)
			if ( doc.getById( 'cke_pastebin' ) )
				return;

			var sel = editor.getSelection();
			var bms = sel.createBookmarks();

			// http://dev.ckeditor.com/ticket/11384. On IE9+ we use native selectionchange (i.e. editor#selectionCheck) to cache the most
			// recent selection which we then lock on editable blur. See selection.js for more info.
			// selectionchange fired before getClipboardDataByPastebin() cached selection
			// before creating bookmark (cached selection will be invalid, because bookmarks modified the DOM),
			// so we need to fire selectionchange one more time, to store current seleciton.
			// Selection will be locked when we focus pastebin.
			if ( CKEDITOR.env.ie )
				sel.root.fire( 'selectionchange' );

			// Create container to paste into.
			// For rich content we prefer to use "body" since it holds
			// the least possibility to be splitted by pasted content, while this may
			// breaks the text selection on a frame-less editable, "div" would be
			// the best one in that case.
			// In another case on old IEs moving the selection into a "body" paste bin causes error panic.
			// Body can't be also used for Opera which fills it with <br>
			// what is indistinguishable from pasted <br> (copying <br> in Opera isn't possible,
			// but it can be copied from other browser).
			var pastebin = new CKEDITOR.dom.element(
				( CKEDITOR.env.webkit || editable.is( 'body' ) ) && !CKEDITOR.env.ie ? 'body' : 'div', doc );

			pastebin.setAttributes( {
				id: 'cke_pastebin',
				'data-cke-temp': '1'
			} );

			var containerOffset = 0,
				offsetParent,
				win = doc.getWindow();

			if ( CKEDITOR.env.webkit ) {
				// It's better to paste close to the real paste destination, so inherited styles
				// (which Webkits will try to compensate by styling span) differs less from the destination's one.
				editable.append( pastebin );
				// Style pastebin like .cke_editable, to minimize differences between origin and destination. (http://dev.ckeditor.com/ticket/9754)
				pastebin.addClass( 'cke_editable' );

				// Compensate position of offsetParent.
				if ( !editable.is( 'body' ) ) {
					// We're not able to get offsetParent from pastebin (body element), so check whether
					// its parent (editable) is positioned.
					if ( editable.getComputedStyle( 'position' ) != 'static' )
						offsetParent = editable;
					// And if not - safely get offsetParent from editable.
					else
						offsetParent = CKEDITOR.dom.element.get( editable.$.offsetParent );

					containerOffset = offsetParent.getDocumentPosition().y;
				}
			} else {
				// Opera and IE doesn't allow to append to html element.
				editable.getAscendant( CKEDITOR.env.ie ? 'body' : 'html', 1 ).append( pastebin );
			}

			pastebin.setStyles( {
				position: 'absolute',
				// Position the bin at the top (+10 for safety) of viewport to avoid any subsequent document scroll.
				top: ( win.getScrollPosition().y - containerOffset + 10 ) + 'px',
				width: '1px',
				// Caret has to fit in that height, otherwise browsers like Chrome & Opera will scroll window to show it.
				// Set height equal to viewport's height - 20px (safety gaps), minimum 1px.
				height: Math.max( 1, win.getViewPaneSize().height - 20 ) + 'px',
				overflow: 'hidden',
				// Reset styles that can mess up pastebin position.
				margin: 0,
				padding: 0
			} );

			// Paste fails in Safari when the body tag has 'user-select: none'. (http://dev.ckeditor.com/ticket/12506)
			if ( CKEDITOR.env.safari )
				pastebin.setStyles( CKEDITOR.tools.cssVendorPrefix( 'user-select', 'text' ) );

			// Check if the paste bin now establishes new editing host.
			var isEditingHost = pastebin.getParent().isReadOnly();

			if ( isEditingHost ) {
				// Hide the paste bin.
				pastebin.setOpacity( 0 );
				// And make it editable.
				pastebin.setAttribute( 'contenteditable', true );
			}
			// Transparency is not enough since positioned non-editing host always shows
			// resize handler, pull it off the screen instead.
			else {
				pastebin.setStyle( editor.config.contentsLangDirection == 'ltr' ? 'left' : 'right', '-10000px' );
			}

			editor.on( 'selectionChange', cancel, null, null, 0 );

			// Webkit fill fire blur on editable when moving selection to
			// pastebin (if body is used). Cancel it because it causes incorrect
			// selection lock in case of inline editor (http://dev.ckeditor.com/ticket/10644).
			// The same seems to apply to Firefox (http://dev.ckeditor.com/ticket/10787).
			if ( CKEDITOR.env.webkit || CKEDITOR.env.gecko )
				blurListener = editable.once( 'blur', cancel, null, null, -100 );

			// Temporarily move selection to the pastebin.
			isEditingHost && pastebin.focus();
			var range = new CKEDITOR.dom.range( pastebin );
			range.selectNodeContents( pastebin );
			var selPastebin = range.select();

			// If non-native paste is executed, IE will open security alert and blur editable.
			// Editable will then lock selection inside itself and after accepting security alert
			// this selection will be restored. We overwrite stored selection, so it's restored
			// in pastebin. (http://dev.ckeditor.com/ticket/9552)
			if ( CKEDITOR.env.ie ) {
				blurListener = editable.once( 'blur', function() {
					editor.lockSelection( selPastebin );
				} );
			}

			var scrollTop = CKEDITOR.document.getWindow().getScrollPosition().y;

			// Wait a while and grab the pasted contents.
			setTimeout( function() {
				// Restore main window's scroll position which could have been changed
				// by browser in cases described in http://dev.ckeditor.com/ticket/9771.
				if ( CKEDITOR.env.webkit )
					CKEDITOR.document.getBody().$.scrollTop = scrollTop;

				// Blur will be fired only on non-native paste. In other case manually remove listener.
				blurListener && blurListener.removeListener();

				// Restore properly the document focus. (http://dev.ckeditor.com/ticket/8849)
				if ( CKEDITOR.env.ie )
					editable.focus();

				// IE7: selection must go before removing pastebin. (http://dev.ckeditor.com/ticket/8691)
				sel.selectBookmarks( bms );
				pastebin.remove();

				// Grab the HTML contents.
				// We need to look for a apple style wrapper on webkit it also adds
				// a div wrapper if you copy/paste the body of the editor.
				// Remove hidden div and restore selection.
				var bogusSpan;
				if ( CKEDITOR.env.webkit && ( bogusSpan = pastebin.getFirst() ) && ( bogusSpan.is && bogusSpan.hasClass( 'Apple-style-span' ) ) )
					pastebin = bogusSpan;

				editor.removeListener( 'selectionChange', cancel );
				callback( pastebin.getHtml() );
			}, 0 );
		}

		// Try to get content directly on IE from clipboard, without native event
		// being fired before. In other words - synthetically get clipboard data, if it's possible.
		// mainPasteEvent will be fired, so if forced native paste:
		// * worked, getClipboardDataByPastebin will grab it,
		// * didn't work, dataValue and dataTransfer will be empty and editor#paste won't be fired.
		// Clipboard data can be accessed directly only on IEs older than Edge.
		// On other browsers we should fire beforePaste event and return false.
		function getClipboardDataDirectly() {
			if ( clipboard.mainPasteEvent == 'paste' ) {
				editor.fire( 'beforePaste', { type: 'auto', method: 'paste' } );
				return false;
			}

			// Prevent IE from pasting at the begining of the document.
			editor.focus();

			// Command will be handled by 'beforepaste', but as
			// execIECommand( 'paste' ) will fire also 'paste' event
			// we're canceling it.
			preventPasteEventNow();

			// http://dev.ckeditor.com/ticket/9247: Lock focus to prevent IE from hiding toolbar for inline editor.
			var focusManager = editor.focusManager;
			focusManager.lock();

			if ( editor.editable().fire( clipboard.mainPasteEvent ) && !execIECommand( 'paste' ) ) {
				focusManager.unlock();
				return false;
			}
			focusManager.unlock();

			return true;
		}

		// Listens for some clipboard related keystrokes, so they get customized.
		// Needs to be bind to keydown event.
		function onKey( event ) {
			if ( editor.mode != 'wysiwyg' )
				return;

			switch ( event.data.keyCode ) {
				// Paste
				case CKEDITOR.CTRL + 86: // CTRL+V
				case CKEDITOR.SHIFT + 45: // SHIFT+INS
					var editable = editor.editable();

					// Cancel 'paste' event because ctrl+v is for IE handled
					// by 'beforepaste'.
					preventPasteEventNow();

					// Simulate 'beforepaste' event for all browsers using 'paste' as main event.
					if ( clipboard.mainPasteEvent == 'paste' ) {
						editable.fire( 'beforepaste' );
					}

					return;

					// Cut
				case CKEDITOR.CTRL + 88: // CTRL+X
				case CKEDITOR.SHIFT + 46: // SHIFT+DEL
					// Save Undo snapshot.
					editor.fire( 'saveSnapshot' ); // Save before cut
					setTimeout( function() {
						editor.fire( 'saveSnapshot' ); // Save after cut
					}, 50 ); // OSX is slow (http://dev.ckeditor.com/ticket/11416).
			}
		}

		function pasteDataFromClipboard( evt ) {
			// Default type is 'auto', but can be changed by beforePaste listeners.
			var eventData = {
					type: 'auto',
					method: 'paste',
					dataTransfer: clipboard.initPasteDataTransfer( evt )
				};

			eventData.dataTransfer.cacheData();

			// Fire 'beforePaste' event so clipboard flavor get customized by other plugins.
			// If 'beforePaste' is canceled continue executing getClipboardDataByPastebin and then do nothing
			// (do not fire 'paste', 'afterPaste' events). This way we can grab all - synthetically
			// and natively pasted content and prevent its insertion into editor
			// after canceling 'beforePaste' event.
			var beforePasteNotCanceled = editor.fire( 'beforePaste', eventData ) !== false;

			// Do not use paste bin if the browser let us get HTML or files from dataTranfer.
			if ( beforePasteNotCanceled && clipboard.canClipboardApiBeTrusted( eventData.dataTransfer, editor ) ) {
				evt.data.preventDefault();
				setTimeout( function() {
					firePasteEvents( editor, eventData );
				}, 0 );
			} else {
				getClipboardDataByPastebin( evt, function( data ) {
					// Clean up.
					eventData.dataValue = data.replace( /<span[^>]+data-cke-bookmark[^<]*?<\/span>/ig, '' );

					// Fire remaining events (without beforePaste)
					beforePasteNotCanceled && firePasteEvents( editor, eventData );
				} );
			}
		}

		function setToolbarStates() {
			if ( editor.mode != 'wysiwyg' )
				return;

			var pasteState = stateFromNamedCommand( 'paste' );

			editor.getCommand( 'cut' ).setState( stateFromNamedCommand( 'cut' ) );
			editor.getCommand( 'copy' ).setState( stateFromNamedCommand( 'copy' ) );
			editor.getCommand( 'paste' ).setState( pasteState );
			editor.fire( 'pasteState', pasteState );
		}

		function stateFromNamedCommand( command ) {
			if ( inReadOnly && command in { paste: 1, cut: 1 } )
				return CKEDITOR.TRISTATE_DISABLED;

			if ( command == 'paste' )
				return CKEDITOR.TRISTATE_OFF;

			// Cut, copy - check if the selection is not empty.
			var sel = editor.getSelection(),
				ranges = sel.getRanges(),
				selectionIsEmpty = sel.getType() == CKEDITOR.SELECTION_NONE || ( ranges.length == 1 && ranges[ 0 ].collapsed );

			return selectionIsEmpty ? CKEDITOR.TRISTATE_DISABLED : CKEDITOR.TRISTATE_OFF;
		}
	}

	// Returns:
	// * 'htmlifiedtext' if content looks like transformed by browser from plain text.
	//		See clipboard/paste.html TCs for more info.
	// * 'html' if it is not 'htmlifiedtext'.
	function recogniseContentType( data ) {
		if ( CKEDITOR.env.webkit ) {
			// Plain text or ( <div><br></div> and text inside <div> ).
			if ( !data.match( /^[^<]*$/g ) && !data.match( /^(<div><br( ?\/)?><\/div>|<div>[^<]*<\/div>)*$/gi ) )
				return 'html';
		} else if ( CKEDITOR.env.ie ) {
			// Text and <br> or ( text and <br> in <p> - paragraphs can be separated by new \r\n ).
			if ( !data.match( /^([^<]|<br( ?\/)?>)*$/gi ) && !data.match( /^(<p>([^<]|<br( ?\/)?>)*<\/p>|(\r\n))*$/gi ) )
				return 'html';
		} else if ( CKEDITOR.env.gecko ) {
			// Text or <br>.
			if ( !data.match( /^([^<]|<br( ?\/)?>)*$/gi ) )
				return 'html';
		} else {
			return 'html';
		}

		return 'htmlifiedtext';
	}

	// This function transforms what browsers produce when
	// pasting plain text into editable element (see clipboard/paste.html TCs
	// for more info) into correct HTML (similar to that produced by text2Html).
	function htmlifiedTextHtmlification( config, data ) {
		function repeatParagraphs( repeats ) {
			// Repeat blocks floor((n+1)/2) times.
			// Even number of repeats - add <br> at the beginning of last <p>.
			return CKEDITOR.tools.repeat( '</p><p>', ~~( repeats / 2 ) ) + ( repeats % 2 == 1 ? '<br>' : '' );
		}

			// Replace adjacent white-spaces (EOLs too - Fx sometimes keeps them) with one space.
		data = data.replace( /\s+/g, ' ' )
			// Remove spaces from between tags.
			.replace( /> +</g, '><' )
			// Normalize XHTML syntax and upper cased <br> tags.
			.replace( /<br ?\/>/gi, '<br>' );

		// IE - lower cased tags.
		data = data.replace( /<\/?[A-Z]+>/g, function( match ) {
			return match.toLowerCase();
		} );

		// Don't touch single lines (no <br|p|div>) - nothing to do here.
		if ( data.match( /^[^<]$/ ) )
			return data;

		// Webkit.
		if ( CKEDITOR.env.webkit && data.indexOf( '<div>' ) > -1 ) {
				// One line break at the beginning - insert <br>
			data = data.replace( /^(<div>(<br>|)<\/div>)(?!$|(<div>(<br>|)<\/div>))/g, '<br>' )
				// Two or more - reduce number of new lines by one.
				.replace( /^(<div>(<br>|)<\/div>){2}(?!$)/g, '<div></div>' );

			// Two line breaks create one paragraph in Webkit.
			if ( data.match( /<div>(<br>|)<\/div>/ ) ) {
				data = '<p>' + data.replace( /(<div>(<br>|)<\/div>)+/g, function( match ) {
					return repeatParagraphs( match.split( '</div><div>' ).length + 1 );
				} ) + '</p>';
			}

			// One line break create br.
			data = data.replace( /<\/div><div>/g, '<br>' );

			// Remove remaining divs.
			data = data.replace( /<\/?div>/g, '' );
		}

		// Opera and Firefox and enterMode != BR.
		if ( CKEDITOR.env.gecko && config.enterMode != CKEDITOR.ENTER_BR ) {
			// Remove bogus <br> - Fx generates two <brs> for one line break.
			// For two line breaks it still produces two <brs>, but it's better to ignore this case than the first one.
			if ( CKEDITOR.env.gecko )
				data = data.replace( /^<br><br>$/, '<br>' );

			// This line satisfy edge case when for Opera we have two line breaks
			//data = data.replace( /)

			if ( data.indexOf( '<br><br>' ) > -1 ) {
				// Two line breaks create one paragraph, three - 2, four - 3, etc.
				data = '<p>' + data.replace( /(<br>){2,}/g, function( match ) {
					return repeatParagraphs( match.length / 4 );
				} ) + '</p>';
			}
		}

		return switchEnterMode( config, data );
	}

	function filtersFactoryFactory() {
		var filters = {};

		function setUpTags() {
			var tags = {};

			for ( var tag in CKEDITOR.dtd ) {
				if ( tag.charAt( 0 ) != '$' && tag != 'div' && tag != 'span' ) {
					tags[ tag ] = 1;
				}
			}

			return tags;
		}

		function createSemanticContentFilter() {
			var filter = new CKEDITOR.filter();

			filter.allow( {
				$1: {
					elements: setUpTags(),
					attributes: true,
					styles: false,
					classes: false
				}
			} );

			return filter;
		}

		return {
			get: function( type ) {
				if ( type == 'plain-text' ) {
					// Does this look confusing to you? Did we forget about enter mode?
					// It is a trick that let's us creating one filter for edidtor, regardless of its
					// activeEnterMode (which as the name indicates can change during runtime).
					//
					// How does it work?
					// The active enter mode is passed to the filter.applyTo method.
					// The filter first marks all elements except <br> as disallowed and then tries to remove
					// them. However, it cannot remove e.g. a <p> element completely, because it's a basic structural element,
					// so it tries to replace it with an element created based on the active enter mode, eventually doing nothing.
					//
					// Now you can sleep well.
					return filters.plainText || ( filters.plainText = new CKEDITOR.filter( 'br' ) );
				} else if ( type == 'semantic-content' ) {
					return filters.semanticContent || ( filters.semanticContent = createSemanticContentFilter() );
				} else if ( type ) {
					// Create filter based on rules (string or object).
					return new CKEDITOR.filter( type );
				}

				return null;
			}
		};
	}

	function filterContent( editor, data, filter ) {
		var fragment = CKEDITOR.htmlParser.fragment.fromHtml( data ),
			writer = new CKEDITOR.htmlParser.basicWriter();

		filter.applyTo( fragment, true, false, editor.activeEnterMode );
		fragment.writeHtml( writer );

		return writer.getHtml();
	}

	function switchEnterMode( config, data ) {
		if ( config.enterMode == CKEDITOR.ENTER_BR ) {
			data = data.replace( /(<\/p><p>)+/g, function( match ) {
				return CKEDITOR.tools.repeat( '<br>', match.length / 7 * 2 );
			} ).replace( /<\/?p>/g, '' );
		} else if ( config.enterMode == CKEDITOR.ENTER_DIV ) {
			data = data.replace( /<(\/)?p>/g, '<$1div>' );
		}

		return data;
	}

	function preventDefaultSetDropEffectToNone( evt ) {
		evt.data.preventDefault();
		evt.data.$.dataTransfer.dropEffect = 'none';
	}

	function initDragDrop( editor ) {
		var clipboard = CKEDITOR.plugins.clipboard;

		editor.on( 'contentDom', function() {
			var editable = editor.editable(),
				dropTarget = CKEDITOR.plugins.clipboard.getDropTarget( editor ),
				top = editor.ui.space( 'top' ),
				bottom = editor.ui.space( 'bottom' );

			// -------------- DRAGOVER TOP & BOTTOM --------------

			// Not allowing dragging on toolbar and bottom (http://dev.ckeditor.com/ticket/12613).
			clipboard.preventDefaultDropOnElement( top );
			clipboard.preventDefaultDropOnElement( bottom );

			// -------------- DRAGSTART --------------
			// Listed on dragstart to mark internal and cross-editor drag & drop
			// and save range and selected HTML.

			editable.attachListener( dropTarget, 'dragstart', fireDragEvent );

			// Make sure to reset data transfer (in case dragend was not called or was canceled).
			editable.attachListener( editor, 'dragstart', clipboard.resetDragDataTransfer, clipboard, null, 1 );

			// Create a dataTransfer object and save it globally.
			editable.attachListener( editor, 'dragstart', function( evt ) {
				clipboard.initDragDataTransfer( evt, editor );
			}, null, null, 2 );

			editable.attachListener( editor, 'dragstart', function() {
				// Save drag range globally for cross editor D&D.
				var dragRange = clipboard.dragRange = editor.getSelection().getRanges()[ 0 ];

				// Store number of children, so we can later tell if any text node was split on drop. (http://dev.ckeditor.com/ticket/13011, http://dev.ckeditor.com/ticket/13447)
				if ( CKEDITOR.env.ie && CKEDITOR.env.version < 10 ) {
					clipboard.dragStartContainerChildCount = dragRange ? getContainerChildCount( dragRange.startContainer ) : null;
					clipboard.dragEndContainerChildCount = dragRange ? getContainerChildCount( dragRange.endContainer ) : null;
				}
			}, null, null, 100 );

			// -------------- DRAGEND --------------
			// Clean up on dragend.

			editable.attachListener( dropTarget, 'dragend', fireDragEvent );

			// Init data transfer if someone wants to use it in dragend.
			editable.attachListener( editor, 'dragend', clipboard.initDragDataTransfer, clipboard, null, 1 );

			// When drag & drop is done we need to reset dataTransfer so the future
			// external drop will be not recognize as internal.
			editable.attachListener( editor, 'dragend', clipboard.resetDragDataTransfer, clipboard, null, 100 );

			// -------------- DRAGOVER --------------
			// We need to call preventDefault on dragover because otherwise if
			// we drop image it will overwrite document.

			editable.attachListener( dropTarget, 'dragover', function( evt ) {
				// Edge requires this handler to have `preventDefault()` regardless of the situation.
				if ( CKEDITOR.env.edge ) {
					evt.data.preventDefault();
					return;
				}

				var target = evt.data.getTarget();

				// Prevent reloading page when dragging image on empty document (http://dev.ckeditor.com/ticket/12619).
				if ( target && target.is && target.is( 'html' ) ) {
					evt.data.preventDefault();
					return;
				}

				// If we do not prevent default dragover on IE the file path
				// will be loaded and we will lose content. On the other hand
				// if we prevent it the cursor will not we shown, so we prevent
				// dragover only on IE, on versions which support file API and only
				// if the event contains files.
				if ( CKEDITOR.env.ie &&
					CKEDITOR.plugins.clipboard.isFileApiSupported &&
					evt.data.$.dataTransfer.types.contains( 'Files' ) ) {
					evt.data.preventDefault();
				}
			} );

			// -------------- DROP --------------

			editable.attachListener( dropTarget, 'drop', function( evt ) {
				// Do nothing if event was already prevented. (http://dev.ckeditor.com/ticket/13879)
				if ( evt.data.$.defaultPrevented ) {
					return;
				}

				// Cancel native drop.
				evt.data.preventDefault();

				var target = evt.data.getTarget(),
					readOnly = target.isReadOnly();

				// Do nothing if drop on non editable element (http://dev.ckeditor.com/ticket/13015).
				// The <html> tag isn't editable (body is), but we want to allow drop on it
				// (so it is possible to drop below editor contents).
				if ( readOnly && !( target.type == CKEDITOR.NODE_ELEMENT && target.is( 'html' ) ) ) {
					return;
				}

				// Getting drop position is one of the most complex parts.
				var dropRange = clipboard.getRangeAtDropPosition( evt, editor ),
					dragRange = clipboard.dragRange;

				// Do nothing if it was not possible to get drop range.
				if ( !dropRange ) {
					return;
				}

				// Fire drop.
				fireDragEvent( evt, dragRange, dropRange  );
			}, null, null, 9999 );

			// Create dataTransfer or get it, if it was created before.
			editable.attachListener( editor, 'drop', clipboard.initDragDataTransfer, clipboard, null, 1 );

			// Execute drop action, fire paste.
			editable.attachListener( editor, 'drop', function( evt ) {
				var data = evt.data;

				if ( !data ) {
					return;
				}

				// Let user modify drag and drop range.
				var dropRange = data.dropRange,
					dragRange = data.dragRange,
					dataTransfer = data.dataTransfer;

				if ( dataTransfer.getTransferType( editor ) == CKEDITOR.DATA_TRANSFER_INTERNAL ) {
					// Execute drop with a timeout because otherwise selection, after drop,
					// on IE is in the drag position, instead of drop position.
					setTimeout( function() {
						clipboard.internalDrop( dragRange, dropRange, dataTransfer, editor );
					}, 0 );
				} else if ( dataTransfer.getTransferType( editor ) == CKEDITOR.DATA_TRANSFER_CROSS_EDITORS ) {
					crossEditorDrop( dragRange, dropRange, dataTransfer );
				} else {
					externalDrop( dropRange, dataTransfer );
				}
			}, null, null, 9999 );

			// Cross editor drag and drop (drag in one Editor and drop in the other).
			function crossEditorDrop( dragRange, dropRange, dataTransfer ) {
				// Paste event should be fired before delete contents because otherwise
				// Chrome have a problem with drop range (Chrome split the drop
				// range container so the offset is bigger then container length).
				dropRange.select();
				firePasteEvents( editor, { dataTransfer: dataTransfer, method: 'drop' }, 1 );

				// Remove dragged content and make a snapshot.
				dataTransfer.sourceEditor.fire( 'saveSnapshot' );

				dataTransfer.sourceEditor.editable().extractHtmlFromRange( dragRange );

				// Make some selection before saving snapshot, otherwise error will be thrown, because
				// there will be no valid selection after content is removed.
				dataTransfer.sourceEditor.getSelection().selectRanges( [ dragRange ] );
				dataTransfer.sourceEditor.fire( 'saveSnapshot' );
			}

			// Drop from external source.
			function externalDrop( dropRange, dataTransfer ) {
				// Paste content into the drop position.
				dropRange.select();

				firePasteEvents( editor, { dataTransfer: dataTransfer, method: 'drop' }, 1 );

				// Usually we reset DataTranfer on dragend,
				// but dragend is called on the same element as dragstart
				// so it will not be called on on external drop.
				clipboard.resetDragDataTransfer();
			}

			// Fire drag/drop events (dragstart, dragend, drop).
			function fireDragEvent( evt, dragRange, dropRange ) {
				var eventData = {
						$: evt.data.$,
						target: evt.data.getTarget()
					};

				if ( dragRange ) {
					eventData.dragRange = dragRange;
				}
				if ( dropRange ) {
					eventData.dropRange = dropRange;
				}

				if ( editor.fire( evt.name, eventData ) === false ) {
					evt.data.preventDefault();
				}
			}

			function getContainerChildCount( container ) {
				if ( container.type != CKEDITOR.NODE_ELEMENT ) {
					container = container.getParent();
				}

				return container.getChildCount();
			}
		} );
	}

	/**
	 * @singleton
	 * @class CKEDITOR.plugins.clipboard
	 */
	CKEDITOR.plugins.clipboard = {
		/**
		 * True if the environment allows to set data on copy or cut manually. This value is false in IE, because this browser
		 * shows the security dialog window when the script tries to set clipboard data and on iOS, because custom data is
		 * not saved to clipboard there.
		 *
		 * @since 4.5
		 * @readonly
		 * @property {Boolean}
		 */
		isCustomCopyCutSupported: !CKEDITOR.env.ie && !CKEDITOR.env.iOS,

		/**
		 * True if the environment supports MIME types and custom data types in dataTransfer/cliboardData getData/setData methods.
		 *
		 * @since 4.5
		 * @readonly
		 * @property {Boolean}
		 */
		isCustomDataTypesSupported: !CKEDITOR.env.ie,

		/**
		 * True if the environment supports File API.
		 *
		 * @since 4.5
		 * @readonly
		 * @property {Boolean}
		 */
		isFileApiSupported: !CKEDITOR.env.ie || CKEDITOR.env.version > 9,

		/**
		 * Main native paste event editable should listen to.
		 *
		 * **Note:** Safari does not like the {@link CKEDITOR.editor#beforePaste} event &mdash; it sometimes does not
		 * handle <kbd>Ctrl+C</kbd> properly. This is probably caused by some race condition between events.
		 * Chrome, Firefox and Edge work well with both events, so it is better to use {@link CKEDITOR.editor#paste}
		 * which will handle pasting from e.g. browsers' menu bars.
		 * IE7/8 does not like the {@link CKEDITOR.editor#paste} event for which it is throwing random errors.
		 *
		 * @since 4.5
		 * @readonly
		 * @property {String}
		 */
		mainPasteEvent: ( CKEDITOR.env.ie && !CKEDITOR.env.edge ) ? 'beforepaste' : 'paste',

		/**
		 * Returns `true` if it is expected that a browser provides HTML data through the Clipboard API.
		 * If not, this method returns `false` and as a result CKEditor will use the paste bin. Read more in
		 * the [Clipboard Integration](http://docs.ckeditor.com/#!/guide/dev_clipboard-section-clipboard-api) guide.
		 *
		 * @since 4.5.2
		 * @returns {Boolean}
		 */
		canClipboardApiBeTrusted: function( dataTransfer, editor ) {
			// If it's an internal or cross-editor data transfer, then it means that custom cut/copy/paste support works
			// and that the data were put manually on the data transfer so we can be sure that it's available.
			if ( dataTransfer.getTransferType( editor ) != CKEDITOR.DATA_TRANSFER_EXTERNAL ) {
				return true;
			}

			// In Chrome we can trust Clipboard API, with the exception of Chrome on Android (in both - mobile and desktop modes), where
			// clipboard API is not available so we need to check it (http://dev.ckeditor.com/ticket/13187).
			if ( CKEDITOR.env.chrome && !dataTransfer.isEmpty() ) {
				return true;
			}

			// Because of a Firefox bug HTML data are not available in some cases (e.g. paste from Word), in such cases we
			// need to use the pastebin (http://dev.ckeditor.com/ticket/13528, https://bugzilla.mozilla.org/show_bug.cgi?id=1183686).
			if ( CKEDITOR.env.gecko && ( dataTransfer.getData( 'text/html' ) || dataTransfer.getFilesCount() ) ) {
				return true;
			}

			// Safari fixed clipboard in 10.1 (https://bugs.webkit.org/show_bug.cgi?id=19893) (http://dev.ckeditor.com/ticket/16982).
			// However iOS version still doesn't work well enough (https://bugs.webkit.org/show_bug.cgi?id=19893#c34).
			if ( CKEDITOR.env.safari && CKEDITOR.env.version >= 603 && !CKEDITOR.env.iOS ) {
				return true;
			}

			// In older Safari and IE HTML data is not available though the Clipboard API.
			// In Edge things are a bit messy at the moment -
			// https://connect.microsoft.com/IE/feedback/details/1572456/edge-clipboard-api-text-html-content-messed-up-in-event-clipboarddata
			// It is safer to use the paste bin in unknown cases.
			return false;
		},

		/**
		 * Returns the element that should be used as the target for the drop event.
		 *
		 * @since 4.5
		 * @param {CKEDITOR.editor} editor The editor instance.
		 * @returns {CKEDITOR.dom.domObject} the element that should be used as the target for the drop event.
		 */
		getDropTarget: function( editor ) {
			var editable = editor.editable();

			// http://dev.ckeditor.com/ticket/11123 Firefox needs to listen on document, because otherwise event won't be fired.
			// http://dev.ckeditor.com/ticket/11086 IE8 cannot listen on document.
			if ( ( CKEDITOR.env.ie && CKEDITOR.env.version < 9 ) || editable.isInline() ) {
				return editable;
			} else {
				return editor.document;
			}
		},

		/**
		 * IE 8 & 9 split text node on drop so the first node contains the
		 * text before the drop position and the second contains the rest. If you
		 * drag the content from the same node you will be not be able to get
		 * it (the range becomes invalid), so you need to join them back.
		 *
		 * Note that the first node in IE 8 & 9 is the original node object
		 * but with shortened content.
		 *
		 *		Before:
		 *		  --- Text Node A ----------------------------------
		 *		                                             /\
		 *		                                        Drag position
		 *
		 *		After (IE 8 & 9):
		 *		  --- Text Node A -----  --- Text Node B -----------
		 *		                       /\                    /\
		 *		                  Drop position        Drag position
		 *		                                         (invalid)
		 *
		 *		After (other browsers):
		 *		  --- Text Node A ----------------------------------
		 *		                       /\                    /\
		 *		                  Drop position        Drag position
		 *
		 * **Note:** This function is in the public scope for tests usage only.
		 *
		 * @since 4.5
		 * @private
		 * @param {CKEDITOR.dom.range} dragRange The drag range.
		 * @param {CKEDITOR.dom.range} dropRange The drop range.
		 * @param {Number} preDragStartContainerChildCount The number of children of the drag range start container before the drop.
		 * @param {Number} preDragEndContainerChildCount The number of children of the drag range end container before the drop.
		 */
		fixSplitNodesAfterDrop: function( dragRange, dropRange, preDragStartContainerChildCount, preDragEndContainerChildCount ) {
			var dropContainer = dropRange.startContainer;

			if (
				typeof preDragEndContainerChildCount != 'number' ||
				typeof preDragStartContainerChildCount != 'number'
			) {
				return;
			}

			// We are only concerned about ranges anchored in elements.
			if ( dropContainer.type != CKEDITOR.NODE_ELEMENT ) {
				return;
			}

			if ( handleContainer( dragRange.startContainer, dropContainer, preDragStartContainerChildCount ) ) {
				return;
			}

			if ( handleContainer( dragRange.endContainer, dropContainer, preDragEndContainerChildCount ) ) {
				return;
			}

			function handleContainer( dragContainer, dropContainer, preChildCount ) {
				var dragElement = dragContainer;
				if ( dragElement.type == CKEDITOR.NODE_TEXT ) {
					dragElement = dragContainer.getParent();
				}

				if ( dragElement.equals( dropContainer ) && preChildCount != dropContainer.getChildCount() ) {
					applyFix( dropRange );
					return true;
				}
			}

			function applyFix( dropRange ) {
				var nodeBefore = dropRange.startContainer.getChild( dropRange.startOffset - 1 ),
					nodeAfter = dropRange.startContainer.getChild( dropRange.startOffset );

				if (
					nodeBefore && nodeBefore.type == CKEDITOR.NODE_TEXT &&
					nodeAfter && nodeAfter.type == CKEDITOR.NODE_TEXT
				) {
					var offset = nodeBefore.getLength();

					nodeBefore.setText( nodeBefore.getText() + nodeAfter.getText() );
					nodeAfter.remove();

					dropRange.setStart( nodeBefore, offset );
					dropRange.collapse( true );
				}
			}
		},

		/**
		 * Checks whether turning the drag range into bookmarks will invalidate the drop range.
		 * This usually happens when the drop range shares the container with the drag range and is
		 * located after the drag range, but there are countless edge cases.
		 *
		 * This function is stricly related to {@link #internalDrop} which toggles
		 * order in which it creates bookmarks for both ranges based on a value returned
		 * by this method. In some cases this method returns a value which is not necessarily
		 * true in terms of what it was meant to check, but it is convenient, because
		 * we know how it is interpreted in {@link #internalDrop}, so the correct
		 * behavior of the entire algorithm is assured.
		 *
		 * **Note:** This function is in the public scope for tests usage only.
		 *
		 * @since 4.5
		 * @private
		 * @param {CKEDITOR.dom.range} dragRange The first range to compare.
		 * @param {CKEDITOR.dom.range} dropRange The second range to compare.
		 * @returns {Boolean} `true` if the first range is before the second range.
		 */
		isDropRangeAffectedByDragRange: function( dragRange, dropRange ) {
			var dropContainer = dropRange.startContainer,
				dropOffset = dropRange.endOffset;

			// Both containers are the same and drop offset is at the same position or later.
			// " A L] A " " M A "
			//       ^ ^
			if ( dragRange.endContainer.equals( dropContainer ) && dragRange.endOffset <= dropOffset ) {
				return true;
			}

			// Bookmark for drag start container will mess up with offsets.
			// " O [L A " " M A "
			//           ^       ^
			if (
				dragRange.startContainer.getParent().equals( dropContainer ) &&
				dragRange.startContainer.getIndex() < dropOffset
			) {
				return true;
			}

			// Bookmark for drag end container will mess up with offsets.
			// " O] L A " " M A "
			//           ^       ^
			if (
				dragRange.endContainer.getParent().equals( dropContainer ) &&
				dragRange.endContainer.getIndex() < dropOffset
			) {
				return true;
			}

			return false;
		},

		/**
		 * Internal drag and drop (drag and drop in the same editor instance).
		 *
		 * **Note:** This function is in the public scope for tests usage only.
		 *
		 * @since 4.5
		 * @private
		 * @param {CKEDITOR.dom.range} dragRange The first range to compare.
		 * @param {CKEDITOR.dom.range} dropRange The second range to compare.
		 * @param {CKEDITOR.plugins.clipboard.dataTransfer} dataTransfer
		 * @param {CKEDITOR.editor} editor
		 */
		internalDrop: function( dragRange, dropRange, dataTransfer, editor ) {
			var clipboard = CKEDITOR.plugins.clipboard,
				editable = editor.editable(),
				dragBookmark, dropBookmark, isDropRangeAffected;

			// Save and lock snapshot so there will be only
			// one snapshot for both remove and insert content.
			editor.fire( 'saveSnapshot' );
			editor.fire( 'lockSnapshot', { dontUpdate: 1 } );

			if ( CKEDITOR.env.ie && CKEDITOR.env.version < 10 ) {
				this.fixSplitNodesAfterDrop(
					dragRange,
					dropRange,
					clipboard.dragStartContainerChildCount,
					clipboard.dragEndContainerChildCount
				);
			}

			// Because we manipulate multiple ranges we need to do it carefully,
			// changing one range (event creating a bookmark) may make other invalid.
			// We need to change ranges into bookmarks so we can manipulate them easily in the future.
			// We can change the range which is later in the text before we change the preceding range.
			// We call isDropRangeAffectedByDragRange to test the order of ranges.
			isDropRangeAffected = this.isDropRangeAffectedByDragRange( dragRange, dropRange );
			if ( !isDropRangeAffected ) {
				dragBookmark = dragRange.createBookmark( false );
			}
			dropBookmark = dropRange.clone().createBookmark( false );
			if ( isDropRangeAffected ) {
				dragBookmark = dragRange.createBookmark( false );
			}

			// Check if drop range is inside range.
			// This is an edge case when we drop something on editable's margin/padding.
			// That space is not treated as a part of the range we drag, so it is possible to drop there.
			// When we drop, browser tries to find closest drop position and it finds it inside drag range. (http://dev.ckeditor.com/ticket/13453)
			var startNode = dragBookmark.startNode,
				endNode = dragBookmark.endNode,
				dropNode = dropBookmark.startNode,
				dropInsideDragRange =
					// Must check endNode because dragRange could be collapsed in some edge cases (simulated DnD).
					endNode &&
					( startNode.getPosition( dropNode ) & CKEDITOR.POSITION_PRECEDING ) &&
					( endNode.getPosition( dropNode ) & CKEDITOR.POSITION_FOLLOWING );

			// If the drop range happens to be inside drag range change it's position to the beginning of the drag range.
			if ( dropInsideDragRange ) {
				// We only change position of bookmark span that is connected with dropBookmark.
				// dropRange will be overwritten and set to the dropBookmark later.
				dropNode.insertBefore( startNode );
			}

			// No we can safely delete content for the drag range...
			dragRange = editor.createRange();
			dragRange.moveToBookmark( dragBookmark );
			editable.extractHtmlFromRange( dragRange, 1 );

			// ...and paste content into the drop position.
			dropRange = editor.createRange();
			dropRange.moveToBookmark( dropBookmark );

			// We do not select drop range, because of may be in the place we can not set the selection
			// (e.g. between blocks, in case of block widget D&D). We put range to the paste event instead.
			firePasteEvents( editor, { dataTransfer: dataTransfer, method: 'drop', range: dropRange }, 1 );

			editor.fire( 'unlockSnapshot' );
		},

		/**
		 * Gets the range from the `drop` event.
		 *
		 * @since 4.5
		 * @param {Object} domEvent A native DOM drop event object.
		 * @param {CKEDITOR.editor} editor The source editor instance.
		 * @returns {CKEDITOR.dom.range} range at drop position.
		 */
		getRangeAtDropPosition: function( dropEvt, editor ) {
			var $evt = dropEvt.data.$,
				x = $evt.clientX,
				y = $evt.clientY,
				$range,
				defaultRange = editor.getSelection( true ).getRanges()[ 0 ],
				range = editor.createRange();

			// Make testing possible.
			if ( dropEvt.data.testRange )
				return dropEvt.data.testRange;

			// Webkits.
			if ( document.caretRangeFromPoint && editor.document.$.caretRangeFromPoint( x, y ) ) {
				$range = editor.document.$.caretRangeFromPoint( x, y );
				range.setStart( CKEDITOR.dom.node( $range.startContainer ), $range.startOffset );
				range.collapse( true );
			}
			// FF.
			else if ( $evt.rangeParent ) {
				range.setStart( CKEDITOR.dom.node( $evt.rangeParent ), $evt.rangeOffset );
				range.collapse( true );
			}
			// IEs 9+.
			// We check if editable is focused to make sure that it's an internal DnD. External DnD must use the second
			// mechanism because of http://dev.ckeditor.com/ticket/13472#comment:6.
			else if ( CKEDITOR.env.ie && CKEDITOR.env.version > 8 && defaultRange && editor.editable().hasFocus ) {
				// On IE 9+ range by default is where we expected it.
				// defaultRange may be undefined if dragover was canceled (file drop).
				return defaultRange;
			}
			// IE 8 and all IEs if !defaultRange or external DnD.
			else if ( document.body.createTextRange ) {
				// To use this method we need a focus (which may be somewhere else in case of external drop).
				editor.focus();

				$range = editor.document.getBody().$.createTextRange();
				try {
					var sucess = false;

					// If user drop between text line IEs moveToPoint throws exception:
					//
					//		Lorem ipsum pulvinar purus et euismod
					//
					//		dolor sit amet,| consectetur adipiscing
					//		               *
					//		vestibulum tincidunt augue eget tempus.
					//
					// * - drop position
					// | - expected cursor position
					//
					// So we try to call moveToPoint with +-1px up to +-20px above or
					// below original drop position to find nearest good drop position.
					for ( var i = 0; i < 20 && !sucess; i++ ) {
						if ( !sucess ) {
							try {
								$range.moveToPoint( x, y - i );
								sucess = true;
							} catch ( err ) {
							}
						}
						if ( !sucess ) {
							try {
								$range.moveToPoint( x, y + i );
								sucess = true;
							} catch ( err ) {
							}
						}
					}

					if ( sucess ) {
						var id = 'cke-temp-' + ( new Date() ).getTime();
						$range.pasteHTML( '<span id="' + id + '">\u200b</span>' );

						var span = editor.document.getById( id );
						range.moveToPosition( span, CKEDITOR.POSITION_BEFORE_START );
						span.remove();
					} else {
						// If the fist method does not succeed we might be next to
						// the short element (like header):
						//
						//		Lorem ipsum pulvinar purus et euismod.
						//
						//
						//		SOME HEADER|        *
						//
						//
						//		vestibulum tincidunt augue eget tempus.
						//
						// * - drop position
						// | - expected cursor position
						//
						// In such situation elementFromPoint returns proper element. Using getClientRect
						// it is possible to check if the cursor should be at the beginning or at the end
						// of paragraph.
						var $element = editor.document.$.elementFromPoint( x, y ),
							element = new CKEDITOR.dom.element( $element ),
							rect;

						if ( !element.equals( editor.editable() ) && element.getName() != 'html' ) {
							rect = element.getClientRect();

							if ( x < rect.left ) {
								range.setStartAt( element, CKEDITOR.POSITION_AFTER_START );
								range.collapse( true );
							} else {
								range.setStartAt( element, CKEDITOR.POSITION_BEFORE_END );
								range.collapse( true );
							}
						}
						// If drop happens on no element elementFromPoint returns html or body.
						//
						//		*      |Lorem ipsum pulvinar purus et euismod.
						//
						//		       vestibulum tincidunt augue eget tempus.
						//
						// * - drop position
						// | - expected cursor position
						//
						// In such case we can try to use default selection. If startContainer is not
						// 'editable' element it is probably proper selection.
						else if ( defaultRange && defaultRange.startContainer &&
							!defaultRange.startContainer.equals( editor.editable() ) ) {
							return defaultRange;

						// Otherwise we can not find any drop position and we have to return null
						// and cancel drop event.
						} else {
							return null;
						}

					}
				} catch ( err ) {
					return null;
				}
			} else {
				return null;
			}

			return range;
		},

		/**
		 * This function tries to link the `evt.data.dataTransfer` property of the {@link CKEDITOR.editor#dragstart},
		 * {@link CKEDITOR.editor#dragend} and {@link CKEDITOR.editor#drop} events to a single
		 * {@link CKEDITOR.plugins.clipboard.dataTransfer} object.
		 *
		 * This method is automatically used by the core of the drag and drop functionality and
		 * usually does not have to be called manually when using the drag and drop events.
		 *
		 * This method behaves differently depending on whether the drag and drop events were fired
		 * artificially (to represent a non-native drag and drop) or whether they were caused by the native drag and drop.
		 *
		 * If the native event is not available, then it will create a new {@link CKEDITOR.plugins.clipboard.dataTransfer}
		 * instance (if it does not exist already) and will link it to this and all following event objects until
		 * the {@link #resetDragDataTransfer} method is called. It means that all three drag and drop events must be fired
		 * in order to ensure that the data transfer is bound correctly.
		 *
		 * If the native event is available, then the {@link CKEDITOR.plugins.clipboard.dataTransfer} is identified
		 * by its ID and a new instance is assigned to the `evt.data.dataTransfer` only if the ID changed or
		 * the {@link #resetDragDataTransfer} method was called.
		 *
		 * @since 4.5
		 * @param {CKEDITOR.dom.event} [evt] A drop event object.
		 * @param {CKEDITOR.editor} [sourceEditor] The source editor instance.
		 */
		initDragDataTransfer: function( evt, sourceEditor ) {
			// Create a new dataTransfer object based on the drop event.
			// If this event was used on dragstart to create dataTransfer
			// both dataTransfer objects will have the same id.
			var nativeDataTransfer = evt.data.$ ? evt.data.$.dataTransfer : null,
				dataTransfer = new this.dataTransfer( nativeDataTransfer, sourceEditor );

			if ( !nativeDataTransfer ) {
				// No native event.
				if ( this.dragData ) {
					dataTransfer = this.dragData;
				} else {
					this.dragData = dataTransfer;
				}
			} else {
				// Native event. If there is the same id we will replace dataTransfer with the one
				// created on drag, because it contains drag editor, drag content and so on.
				// Otherwise (in case of drag from external source) we save new object to
				// the global clipboard.dragData.
				if ( this.dragData && dataTransfer.id == this.dragData.id ) {
					dataTransfer = this.dragData;
				} else {
					this.dragData = dataTransfer;
				}
			}

			evt.data.dataTransfer = dataTransfer;
		},

		/**
		 * Removes the global {@link #dragData} so the next call to {@link #initDragDataTransfer}
		 * always creates a new instance of {@link CKEDITOR.plugins.clipboard.dataTransfer}.
		 *
		 * @since 4.5
		 */
		resetDragDataTransfer: function() {
			this.dragData = null;
		},

		/**
		 * Global object storing the data transfer of the current drag and drop operation.
		 * Do not use it directly, use {@link #initDragDataTransfer} and {@link #resetDragDataTransfer}.
		 *
		 * Note: This object is global (meaning that it is not related to a single editor instance)
		 * in order to handle drag and drop from one editor into another.
		 *
		 * @since 4.5
		 * @private
		 * @property {CKEDITOR.plugins.clipboard.dataTransfer} dragData
		 */

		/**
		 * Range object to save the drag range and remove its content after the drop.
		 *
		 * @since 4.5
		 * @private
		 * @property {CKEDITOR.dom.range} dragRange
		 */

		/**
		 * Initializes and links data transfer objects based on the paste event. If the data
		 * transfer object was already initialized on this event, the function will
		 * return that object. In IE it is not possible to link copy/cut and paste events
		 * so the method always returns a new object. The same happens if there is no paste event
		 * passed to the method.
		 *
		 * @since 4.5
		 * @param {CKEDITOR.dom.event} [evt] A paste event object.
		 * @param {CKEDITOR.editor} [sourceEditor] The source editor instance.
		 * @returns {CKEDITOR.plugins.clipboard.dataTransfer} The data transfer object.
		 */
		initPasteDataTransfer: function( evt, sourceEditor ) {
			if ( !this.isCustomCopyCutSupported ) {
				// Edge does not support custom copy/cut, but it have some useful data in the clipboardData (http://dev.ckeditor.com/ticket/13755).
				return new this.dataTransfer( ( CKEDITOR.env.edge && evt && evt.data.$ && evt.data.$.clipboardData ) || null, sourceEditor );
			} else if ( evt && evt.data && evt.data.$ ) {
				var dataTransfer = new this.dataTransfer( evt.data.$.clipboardData, sourceEditor );

				if ( this.copyCutData && dataTransfer.id == this.copyCutData.id ) {
					dataTransfer = this.copyCutData;
					dataTransfer.$ = evt.data.$.clipboardData;
				} else {
					this.copyCutData = dataTransfer;
				}

				return dataTransfer;
			} else {
				return new this.dataTransfer( null, sourceEditor );
			}
		},

		/**
		 * Prevents dropping on the specified element.
		 *
		 * @since 4.5
		 * @param {CKEDITOR.dom.element} element The element on which dropping should be disabled.
		 */
		preventDefaultDropOnElement: function( element ) {
			element && element.on( 'dragover', preventDefaultSetDropEffectToNone );
		}
	};

	// Data type used to link drag and drop events.
	//
	// In IE URL data type is buggie and there is no way to mark drag & drop  without
	// modifying text data (which would be displayed if user drop content to the textarea)
	// so we just read dragged text.
	//
	// In Chrome and Firefox we can use custom data types.
	var clipboardIdDataType = CKEDITOR.plugins.clipboard.isCustomDataTypesSupported ? 'cke/id' : 'Text';
	/**
	 * Facade for the native `dataTransfer`/`clipboadData` object to hide all differences
	 * between browsers.
	 *
	 * @since 4.5
	 * @class CKEDITOR.plugins.clipboard.dataTransfer
	 * @constructor Creates a class instance.
	 * @param {Object} [nativeDataTransfer] A native data transfer object.
	 * @param {CKEDITOR.editor} [editor] The source editor instance. If the editor is defined, dataValue will
	 * be created based on the editor content and the type will be 'html'.
	 */
	CKEDITOR.plugins.clipboard.dataTransfer = function( nativeDataTransfer, editor ) {
		if ( nativeDataTransfer ) {
			this.$ = nativeDataTransfer;
		}

		this._ = {
			metaRegExp: /^<meta.*?>/i,
			bodyRegExp: /<body(?:[\s\S]*?)>([\s\S]*)<\/body>/i,
			fragmentRegExp: /<!--(?:Start|End)Fragment-->/g,

			data: {},
			files: [],

			normalizeType: function( type ) {
				type = type.toLowerCase();

				if ( type == 'text' || type == 'text/plain' ) {
					return 'Text'; // IE support only Text and URL;
				} else if ( type == 'url' ) {
					return 'URL'; // IE support only Text and URL;
				} else {
					return type;
				}
			}
		};

		// Check if ID is already created.
		this.id = this.getData( clipboardIdDataType );

		// If there is no ID we need to create it. Different browsers needs different ID.
		if ( !this.id ) {
			if ( clipboardIdDataType == 'Text' ) {
				// For IE10+ only Text data type is supported and we have to compare dragged
				// and dropped text. If the ID is not set it means that empty string was dragged
				// (ex. image with no alt). We change null to empty string.
				this.id = '';
			} else {
				// String for custom data type.
				this.id = 'cke-' + CKEDITOR.tools.getUniqueId();
			}
		}

		// In IE10+ we can not use any data type besides text, so we do not call setData.
		if ( clipboardIdDataType != 'Text' ) {
			// Try to set ID so it will be passed from the drag to the drop event.
			// On some browsers with some event it is not possible to setData so we
			// need to catch exceptions.
			try {
				this.$.setData( clipboardIdDataType, this.id );
			} catch ( err ) {}
		}

		if ( editor ) {
			this.sourceEditor = editor;

			this.setData( 'text/html', editor.getSelectedHtml( 1 ) );

			// Without setData( 'text', ... ) on dragstart there is no drop event in Safari.
			// Also 'text' data is empty as drop to the textarea does not work if we do not put there text.
			if ( clipboardIdDataType != 'Text' && !this.getData( 'text/plain' ) ) {
				this.setData( 'text/plain', editor.getSelection().getSelectedText() );
			}
		}

		/**
		 * Data transfer ID used to bind all dataTransfer
		 * objects based on the same event (e.g. in drag and drop events).
		 *
		 * @readonly
		 * @property {String} id
		 */

		/**
		 * A native DOM event object.
		 *
		 * @readonly
		 * @property {Object} $
		 */

		/**
		 * Source editor &mdash; the editor where the drag starts.
		 * Might be undefined if the drag starts outside the editor (e.g. when dropping files to the editor).
		 *
		 * @readonly
		 * @property {CKEDITOR.editor} sourceEditor
		 */

		/**
		 * Private properties and methods.
		 *
		 * @private
		 * @property {Object} _
		 */
	};

	/**
	 * Data transfer operation (drag and drop or copy and paste) started and ended in the same
	 * editor instance.
	 *
	 * @since 4.5
	 * @readonly
	 * @property {Number} [=1]
	 * @member CKEDITOR
	 */
	CKEDITOR.DATA_TRANSFER_INTERNAL = 1;

	/**
	 * Data transfer operation (drag and drop or copy and paste) started in one editor
	 * instance and ended in another.
	 *
	 * @since 4.5
	 * @readonly
	 * @property {Number} [=2]
	 * @member CKEDITOR
	 */
	CKEDITOR.DATA_TRANSFER_CROSS_EDITORS = 2;

	/**
	 * Data transfer operation (drag and drop or copy and paste) started outside of the editor.
	 * The source of the data may be a textarea, HTML, another application, etc.
	 *
	 * @since 4.5
	 * @readonly
	 * @property {Number} [=3]
	 * @member CKEDITOR
	 */
	CKEDITOR.DATA_TRANSFER_EXTERNAL = 3;

	CKEDITOR.plugins.clipboard.dataTransfer.prototype = {
		/**
		 * Facade for the native `getData` method.
		 *
		 * @param {String} type The type of data to retrieve.
		 * @param {Boolean} [getNative=false] Indicates if the whole, original content of the dataTransfer should be returned.
		 * Introduced in CKEditor 4.7.0.
		 * @returns {String} type Stored data for the given type or an empty string if the data for that type does not exist.
		 */
		getData: function( type, getNative ) {
			function isEmpty( data ) {
				return data === undefined || data === null || data === '';
			}

			function filterUnwantedCharacters( data ) {
				if ( typeof data !== 'string' ) {
					return data;
				}

				var htmlEnd = data.indexOf( '</html>' );

				if ( htmlEnd !== -1 ) {
					// Just cut everything after `</html>`, so everything after htmlEnd index + length of `</html>`.
					// Required to workaround bug: https://bugs.chromium.org/p/chromium/issues/detail?id=696978
					return data.substring( 0, htmlEnd + 7 );
				}

				return data;
			}

			type = this._.normalizeType( type );

			var data = this._.data[ type ],
				result;

			if ( isEmpty( data ) ) {
				try {
					data = this.$.getData( type );
				} catch ( e ) {}
			}

			if ( isEmpty( data ) ) {
				data = '';
			}

			// Some browsers add <meta http-equiv="content-type" content="text/html; charset=utf-8"> at the begging of the HTML data
			// or surround it with <html><head>...</head><body>(some content)<!--StartFragment--> and <!--EndFragment-->(some content)</body></html>
			// This code removes meta tags and returns only the contents of the <body> element if found. Note that
			// some significant content may be placed outside Start/EndFragment comments so it's kept.
			//
			// See http://dev.ckeditor.com/ticket/13583 for more details.
			// Additionally http://dev.ckeditor.com/ticket/16847 adds a flag allowing to get the whole, original content.
			if ( type == 'text/html' && !getNative ) {
				data = data.replace( this._.metaRegExp, '' );

				// Keep only contents of the <body> element
				result = this._.bodyRegExp.exec( data );
				if ( result && result.length ) {
					data = result[ 1 ];

					// Remove also comments.
					data = data.replace( this._.fragmentRegExp, '' );
				}
			}
			// Firefox on Linux put files paths as a text/plain data if there are files
			// in the dataTransfer object. We need to hide it, because files should be
			// handled on paste only if dataValue is empty.
			else if ( type == 'Text' && CKEDITOR.env.gecko && this.getFilesCount() &&
				data.substring( 0, 7 ) == 'file://' ) {
				data = '';
			}

			return filterUnwantedCharacters( data );
		},

		/**
		 * Facade for the native `setData` method.
		 *
		 * @param {String} type The type of data to retrieve.
		 * @param {String} value The data to add.
		 */
		setData: function( type, value ) {
			type = this._.normalizeType( type );

			this._.data[ type ] = value;

			// There is "Unexpected call to method or property access." error if you try
			// to set data of unsupported type on IE.
			if ( !CKEDITOR.plugins.clipboard.isCustomDataTypesSupported && type != 'URL' && type != 'Text' ) {
				return;
			}

			// If we use the text type to bind the ID, then if someone tries to set the text, we must also
			// update ID accordingly. http://dev.ckeditor.com/ticket/13468.
			if ( clipboardIdDataType == 'Text' && type == 'Text' ) {
				this.id = value;
			}

			try {
				this.$.setData( type, value );
			} catch ( e ) {}
		},

		/**
		 * Gets the data transfer type.
		 *
		 * @param {CKEDITOR.editor} targetEditor The drop/paste target editor instance.
		 * @returns {Number} Possible values: {@link CKEDITOR#DATA_TRANSFER_INTERNAL},
		 * {@link CKEDITOR#DATA_TRANSFER_CROSS_EDITORS}, {@link CKEDITOR#DATA_TRANSFER_EXTERNAL}.
		 */
		getTransferType: function( targetEditor ) {
			if ( !this.sourceEditor ) {
				return CKEDITOR.DATA_TRANSFER_EXTERNAL;
			} else if ( this.sourceEditor == targetEditor ) {
				return CKEDITOR.DATA_TRANSFER_INTERNAL;
			} else {
				return CKEDITOR.DATA_TRANSFER_CROSS_EDITORS;
			}
		},

		/**
		 * Copies the data from the native data transfer to a private cache.
		 * This function is needed because the data from the native data transfer
		 * is available only synchronously to the event listener. It is not possible
		 * to get the data asynchronously, after a timeout, and the {@link CKEDITOR.editor#paste}
		 * event is fired asynchronously &mdash; hence the need for caching the data.
		 */
		cacheData: function() {
			if ( !this.$ ) {
				return;
			}

			var that = this,
				i, file;

			function getAndSetData( type ) {
				type = that._.normalizeType( type );

				var data = that.getData( type, true );
				if ( data ) {
					that._.data[ type ] = data;
				}
			}

			// Copy data.
			if ( CKEDITOR.plugins.clipboard.isCustomDataTypesSupported ) {
				if ( this.$.types ) {
					for ( i = 0; i < this.$.types.length; i++ ) {
						getAndSetData( this.$.types[ i ] );
					}
				}
			} else {
				getAndSetData( 'Text' );
				getAndSetData( 'URL' );
			}

			// Copy files references.
			file = this._getImageFromClipboard();
			if ( ( this.$ && this.$.files ) || file ) {
				this._.files = [];

				// Edge have empty files property with no length (http://dev.ckeditor.com/ticket/13755).
				if ( this.$.files && this.$.files.length ) {
					for ( i = 0; i < this.$.files.length; i++ ) {
						this._.files.push( this.$.files[ i ] );
					}
				}

				// Don't include $.items if both $.files and $.items contains files, because,
				// according to spec and browsers behavior, they contain the same files.
				if ( this._.files.length === 0 && file ) {
					this._.files.push( file );
				}
			}
		},

		/**
		 * Gets the number of files in the dataTransfer object.
		 *
		 * @returns {Number} The number of files.
		 */
		getFilesCount: function() {
			if ( this._.files.length ) {
				return this._.files.length;
			}

			if ( this.$ && this.$.files && this.$.files.length ) {
				return this.$.files.length;
			}

			return this._getImageFromClipboard() ? 1 : 0;
		},

		/**
		 * Gets the file at the index given.
		 *
		 * @param {Number} i Index.
		 * @returns {File} File instance.
		 */
		getFile: function( i ) {
			if ( this._.files.length ) {
				return this._.files[ i ];
			}

			if ( this.$ && this.$.files && this.$.files.length ) {
				return this.$.files[ i ];
			}

			// File or null if the file was not found.
			return i === 0 ? this._getImageFromClipboard() : undefined;
		},

		/**
		 * Checks if the data transfer contains any data.
		 *
		 * @returns {Boolean} `true` if the object contains no data.
		 */
		isEmpty: function() {
			var typesToCheck = {},
				type;

			// If dataTransfer contains files it is not empty.
			if ( this.getFilesCount() ) {
				return false;
			}

			// Add custom types.
			for ( type in this._.data ) {
				typesToCheck[ type ] = 1;
			}

			// Add native types.
			if ( this.$ ) {
				if ( CKEDITOR.plugins.clipboard.isCustomDataTypesSupported ) {
					if ( this.$.types ) {
						for ( var i = 0; i < this.$.types.length; i++ ) {
							typesToCheck[ this.$.types[ i ] ] = 1;
						}
					}
				} else {
					typesToCheck.Text = 1;
					typesToCheck.URL = 1;
				}
			}

			// Remove ID.
			if ( clipboardIdDataType != 'Text' ) {
				typesToCheck[ clipboardIdDataType ] = 0;
			}

			for ( type in typesToCheck ) {
				if ( typesToCheck[ type ] && this.getData( type ) !== '' ) {
					return false;
				}
			}

			return true;
		},

		/**
		 * When the content of the clipboard is pasted in Chrome, the clipboard data object has an empty `files` property,
		 * but it is possible to get the file as `items[0].getAsFile();` (http://dev.ckeditor.com/ticket/12961).
		 *
		 * @private
		 * @returns {File} File instance or `null` if not found.
		 */
		_getImageFromClipboard: function() {
			var file;

			if ( this.$ && this.$.items && this.$.items[ 0 ] ) {
				try {
					file = this.$.items[ 0 ].getAsFile();
					// Duck typing
					if ( file && file.type ) {
						return file;
					}
				} catch ( err ) {
					// noop
				}
			}

			return undefined;
		}
	};
} )();

/**
 * The default content type that is used when pasted data cannot be clearly recognized as HTML or text.
 *
 * For example: `'foo'` may come from a plain text editor or a website. It is not possible to recognize the content
 * type in this case, so the default type will be used. At the same time it is clear that `'<b>example</b> text'` is
 * HTML and its origin is a web page, email or another rich text editor.
 *
 * **Note:** If content type is text, then styles of the paste context are preserved.
 *
 *		CKEDITOR.config.clipboard_defaultContentType = 'text';
 *
 * See also the {@link CKEDITOR.editor#paste} event and read more about the integration with clipboard
 * in the [Clipboard Deep Dive guide](#!/guide/dev_clipboard).
 *
 * @since 4.0
 * @cfg {'html'/'text'} [clipboard_defaultContentType='html']
 * @member CKEDITOR.config
 */

/**
 * Fired after the user initiated a paste action, but before the data is inserted into the editor.
 * The listeners to this event are able to process the content before its insertion into the document.
 *
 * Read more about the integration with clipboard in the [Clipboard Deep Dive guide](#!/guide/dev_clipboard).
 *
 * See also:
 *
 * * the {@link CKEDITOR.config#pasteFilter} option,
 * * the {@link CKEDITOR.editor#drop} event,
 * * the {@link CKEDITOR.plugins.clipboard.dataTransfer} class.
 *
 * @since 3.1
 * @event paste
 * @member CKEDITOR.editor
 * @param {CKEDITOR.editor} editor This editor instance.
 * @param data
 * @param {String} data.type The type of data in `data.dataValue`. Usually `'html'` or `'text'`, but for listeners
 * with a priority smaller than `6` it may also be `'auto'` which means that the content type has not been recognised yet
 * (this will be done by the content type sniffer that listens with priority `6`).
 * @param {String} data.dataValue HTML to be pasted.
 * @param {String} data.method Indicates the data transfer method. It could be drag and drop or copy and paste.
 * Possible values: `'drop'`, `'paste'`. Introduced in CKEditor 4.5.
 * @param {CKEDITOR.plugins.clipboard.dataTransfer} data.dataTransfer Facade for the native dataTransfer object
 * which provides access to various data types and files, and passes some data between linked events
 * (like drag and drop). Introduced in CKEditor 4.5.
 * @param {Boolean} [data.dontFilter=false] Whether the {@link CKEDITOR.editor#pasteFilter paste filter} should not
 * be applied to data. This option has no effect when `data.type` equals `'text'` which means that for instance
 * {@link CKEDITOR.config#forcePasteAsPlainText} has a higher priority. Introduced in CKEditor 4.5.
 */

/**
 * Fired before the {@link #paste} event. Allows to preset data type.
 *
 * **Note:** This event is deprecated. Add a `0` priority listener for the
 * {@link #paste} event instead.
 *
 * @deprecated
 * @event beforePaste
 * @member CKEDITOR.editor
 */

/**
 * Fired after the {@link #paste} event if content was modified. Note that if the paste
 * event does not insert any data, the `afterPaste` event will not be fired.
 *
 * @event afterPaste
 * @member CKEDITOR.editor
 */

/**
 * Facade for the native `drop` event. Fired when the native `drop` event occurs.
 *
 * **Note:** To manipulate dropped data, use the {@link CKEDITOR.editor#paste} event.
 * Use the `drop` event only to control drag and drop operations (e.g. to prevent the ability to drop some content).
 *
 * Read more about integration with drag and drop in the [Clipboard Deep Dive guide](#!/guide/dev_clipboard).
 *
 * See also:
 *
 * * The {@link CKEDITOR.editor#paste} event,
 * * The {@link CKEDITOR.editor#dragstart} and {@link CKEDITOR.editor#dragend} events,
 * * The {@link CKEDITOR.plugins.clipboard.dataTransfer} class.
 *
 * @since 4.5
 * @event drop
 * @member CKEDITOR.editor
 * @param {CKEDITOR.editor} editor This editor instance.
 * @param data
 * @param {Object} data.$ Native drop event.
 * @param {CKEDITOR.dom.node} data.target Drop target.
 * @param {CKEDITOR.plugins.clipboard.dataTransfer} data.dataTransfer DataTransfer facade.
 * @param {CKEDITOR.dom.range} data.dragRange Drag range, lets you manipulate the drag range.
 * Note that dragged HTML is saved as `text/html` data on `dragstart` so if you change the drag range
 * on drop, dropped HTML will not change. You need to change it manually using
 * {@link CKEDITOR.plugins.clipboard.dataTransfer#setData dataTransfer.setData}.
 * @param {CKEDITOR.dom.range} data.dropRange Drop range, lets you manipulate the drop range.
 */

/**
 * Facade for the native `dragstart` event. Fired when the native `dragstart` event occurs.
 *
 * This event can be canceled in order to block the drag start operation. It can also be fired to mimic the start of the drag and drop
 * operation. For instance, the `widget` plugin uses this option to integrate its custom block widget drag and drop with
 * the entire system.
 *
 * Read more about integration with drag and drop in the [Clipboard Deep Dive guide](#!/guide/dev_clipboard).
 *
 * See also:
 *
 * * The {@link CKEDITOR.editor#paste} event,
 * * The {@link CKEDITOR.editor#drop} and {@link CKEDITOR.editor#dragend} events,
 * * The {@link CKEDITOR.plugins.clipboard.dataTransfer} class.
 *
 * @since 4.5
 * @event dragstart
 * @member CKEDITOR.editor
 * @param {CKEDITOR.editor} editor This editor instance.
 * @param data
 * @param {Object} data.$ Native dragstart event.
 * @param {CKEDITOR.dom.node} data.target Drag target.
 * @param {CKEDITOR.plugins.clipboard.dataTransfer} data.dataTransfer DataTransfer facade.
 */

/**
 * Facade for the native `dragend` event. Fired when the native `dragend` event occurs.
 *
 * Read more about integration with drag and drop in the [Clipboard Deep Dive guide](#!/guide/dev_clipboard).
 *
 * See also:
 *
 * * The {@link CKEDITOR.editor#paste} event,
 * * The {@link CKEDITOR.editor#drop} and {@link CKEDITOR.editor#dragend} events,
 * * The {@link CKEDITOR.plugins.clipboard.dataTransfer} class.
 *
 * @since 4.5
 * @event dragend
 * @member CKEDITOR.editor
 * @param {CKEDITOR.editor} editor This editor instance.
 * @param data
 * @param {Object} data.$ Native dragend event.
 * @param {CKEDITOR.dom.node} data.target Drag target.
 * @param {CKEDITOR.plugins.clipboard.dataTransfer} data.dataTransfer DataTransfer facade.
 */

/**
 * Defines a filter which is applied to external data pasted or dropped into the editor. Possible values are:
 *
 * * `'plain-text'` &ndash; Content will be pasted as a plain text.
 * * `'semantic-content'` &ndash; Known tags (except `div`, `span`) with all attributes (except
 * `style` and `class`) will be kept.
 * * `'h1 h2 p div'` &ndash; Custom rules compatible with {@link CKEDITOR.filter}.
 * * `null` &ndash; Content will not be filtered by the paste filter (but it still may be filtered
 * by [Advanced Content Filter](#!/guide/dev_advanced_content_filter)). This value can be used to
 * disable the paste filter in Chrome and Safari, where this option defaults to `'semantic-content'`.
 *
 * Example:
 *
 *		config.pasteFilter = 'plain-text';
 *
 * Custom setting:
 *
 *		config.pasteFilter = 'h1 h2 p ul ol li; img[!src, alt]; a[!href]';
 *
 * Based on this configuration option, a proper {@link CKEDITOR.filter} instance will be defined and assigned to the editor
 * as a {@link CKEDITOR.editor#pasteFilter}. You can tweak the paste filter settings on the fly on this object
 * as well as delete or replace it.
 *
 *		var editor = CKEDITOR.replace( 'editor', {
 *			pasteFilter: 'semantic-content'
 *		} );
 *
 *		editor.on( 'instanceReady', function() {
 *			// The result of this will be that all semantic content will be preserved
 *			// except tables.
 *			editor.pasteFilter.disallow( 'table' );
 *		} );
 *
 * Note that the paste filter is applied only to **external** data. There are three data sources:
 *
 * * copied and pasted in the same editor (internal),
 * * copied from one editor and pasted into another (cross-editor),
 * * coming from all other sources like websites, MS Word, etc. (external).
 *
 * If {@link CKEDITOR.config#allowedContent Advanced Content Filter} is not disabled, then
 * it will also be applied to pasted and dropped data. The paste filter job is to "normalize"
 * external data which often needs to be handled differently than content produced by the editor.
 *
 * This setting defaults to `'semantic-content'` in Chrome, Opera and Safari (all Blink and Webkit based browsers)
 * due to messy HTML which these browsers keep in the clipboard. In other browsers it defaults to `null`.
 *
 * @since 4.5
 * @cfg {String} [pasteFilter='semantic-content' in Chrome and Safari and `null` in other browsers]
 * @member CKEDITOR.config
 */

/**
 * {@link CKEDITOR.filter Content filter} which is used when external data is pasted or dropped into the editor
 * or a forced paste as plain text occurs.
 *
 * This object might be used on the fly to define rules for pasted external content.
 * This object is available and used if the {@link CKEDITOR.plugins.clipboard clipboard} plugin is enabled and
 * {@link CKEDITOR.config#pasteFilter} or {@link CKEDITOR.config#forcePasteAsPlainText} was defined.
 *
 * To enable the filter:
 *
 *		var editor = CKEDITOR.replace( 'editor', {
 *			pasteFilter: 'plain-text'
 *		} );
 *
 * You can also modify the filter on the fly later on:
 *
 *		editor.pasteFilter = new CKEDITOR.filter( 'p h1 h2; a[!href]' );
 *
 * Note that the paste filter is only applied to **external** data. There are three data sources:
 *
 * * copied and pasted in the same editor (internal),
 * * copied from one editor and pasted into another (cross-editor),
 * * coming from all other sources like websites, MS Word, etc. (external).
 *
 * If {@link CKEDITOR.config#allowedContent Advanced Content Filter} is not disabled, then
 * it will also be applied to pasted and dropped data. The paste filter job is to "normalize"
 * external data which often needs to be handled differently than content produced by the editor.
 *
 * @since 4.5
 * @readonly
 * @property {CKEDITOR.filter} [pasteFilter]
 * @member CKEDITOR.editor
 */

/**
 * Duration of the notification displayed after pasting was blocked by the browser.
 *
 * @since 4.7.0
 * @cfg {Number} [clipboard_notificationDuration=10000]
 * @member CKEDITOR.config
 */
CKEDITOR.config.clipboard_notificationDuration = 10000;
