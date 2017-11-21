/**
 * @license Copyright (c) 2003-2017, CKSource - Frederico Knabben. All rights reserved.
 * For licensing, see LICENSE.md or http://ckeditor.com/license
 */

'use strict';

( function() {
	var validLinkRegExp = /^<a[^>]+href="([^"]+)"[^>]*>([^<]+)<\/a>$/i;

	CKEDITOR.plugins.add( 'autoembed', {
		requires: 'autolink,undo',
		lang: 'az,ca,cs,de,de-ch,en,eo,es,es-mx,eu,fr,gl,hr,hu,it,ja,km,ko,ku,mk,nb,nl,oc,pl,pt,pt-br,ru,sk,sv,tr,ug,uk,zh,zh-cn', // %REMOVE_LINE_CORE%
		init: function( editor ) {
			var currentId = 1,
				embedCandidatePasted;

			editor.on( 'paste', function( evt ) {
				if ( evt.data.dataTransfer.getTransferType( editor ) == CKEDITOR.DATA_TRANSFER_INTERNAL ) {
					embedCandidatePasted = 0;
					return;
				}

				var match = evt.data.dataValue.match( validLinkRegExp );

				embedCandidatePasted = match != null && decodeURI( match[ 1 ] ) == decodeURI( match[ 2 ] );

				// Expecting exactly one <a> tag spanning the whole pasted content.
				// The tag has to have same href as content.
				if ( embedCandidatePasted ) {
					evt.data.dataValue = '<a data-cke-autoembed="' + ( ++currentId ) + '"' + evt.data.dataValue.substr( 2 );
				}
			}, null, null, 20 ); // Execute after autolink.

			editor.on( 'afterPaste', function() {
				// If one pasted an embeddable link and then undone the action, the link in the content holds the
				// data-cke-autoembed attribute and may be embedded on *any* successive paste.
				// This check ensures that autoEmbedLink is called only if afterPaste is fired *right after*
				// embeddable link got into the content. (http://dev.ckeditor.com/ticket/13532)
				if ( embedCandidatePasted ) {
					autoEmbedLink( editor, currentId );
				}
			} );
		}
	} );

	function autoEmbedLink( editor, id ) {
		var anchor = editor.editable().findOne( 'a[data-cke-autoembed="' + id + '"]' ),
			lang = editor.lang.autoembed,
			notification;

		if ( !anchor || !anchor.data( 'cke-saved-href' ) ) {
			return;
		}

		var href = anchor.data( 'cke-saved-href' ),
			widgetDef = CKEDITOR.plugins.autoEmbed.getWidgetDefinition( editor, href );

		if ( !widgetDef ) {
			CKEDITOR.warn( 'autoembed-no-widget-def' );
			return;
		}

			// TODO Move this to a method in the widget plugin. http://dev.ckeditor.com/ticket/13408
		var defaults = typeof widgetDef.defaults == 'function' ? widgetDef.defaults() : widgetDef.defaults,
			element = CKEDITOR.dom.element.createFromHtml( widgetDef.template.output( defaults ) ),
			instance,
			wrapper = editor.widgets.wrapElement( element, widgetDef.name ),
			temp = new CKEDITOR.dom.documentFragment( wrapper.getDocument() );

		temp.append( wrapper );
		instance = editor.widgets.initOn( element, widgetDef );

		if ( !instance ) {
			finalizeCreation();
			return;
		}

		notification = editor.showNotification( lang.embeddingInProgress, 'info' );
		instance.loadContent( href, {
			noNotifications: true,
			callback: function() {
					// DOM might be invalidated in the meantime, so find the anchor again.
				var anchor = editor.editable().findOne( 'a[data-cke-autoembed="' + id + '"]' );

				// Anchor might be removed in the meantime.
				if ( anchor ) {
					var selection = editor.getSelection(),
						insertRange = editor.createRange(),
						editable = editor.editable();

					// Save the changes in editor contents that happened *after* the link was pasted
					// but before it gets embedded (i.e. user pasted and typed).
					editor.fire( 'saveSnapshot' );

					// Lock snapshot so we don't make unnecessary undo steps in
					// editable.insertElement() below, which would include bookmarks. (http://dev.ckeditor.com/ticket/13429)
					editor.fire( 'lockSnapshot', { dontUpdate: true } );

					// Bookmark current selection. (http://dev.ckeditor.com/ticket/13429)
					var bookmark = selection.createBookmarks( false )[ 0 ],
						startNode = bookmark.startNode,
						endNode = bookmark.endNode || startNode;

					// When url is pasted, IE8 sets the caret after <a> element instead of inside it.
					// So, if user hasn't changed selection, bookmark is inserted right after <a>.
					// Then, after pasting embedded content, bookmark is still in DOM but it is
					// inside the original element. After selection recreation it would end up before widget:
					// <p>A <a /><bm /></p><p>B</p>  -->  <p>A <bm /></p><widget /><p>B</p>  -->  <p>A ^</p><widget /><p>B</p>
					// We have to fix this IE8 behavior so it is the same as on other browsers.
					if ( CKEDITOR.env.ie && CKEDITOR.env.version < 9 && !bookmark.endNode && startNode.equals( anchor.getNext() ) ) {
						anchor.append( startNode );
					}

					insertRange.setStartBefore( anchor );
					insertRange.setEndAfter( anchor );

					editable.insertElement( wrapper, insertRange );

					// If both bookmarks are still in DOM, it means that selection was not inside
					// an anchor that got substituted. We can safely recreate that selection. (http://dev.ckeditor.com/ticket/13429)
					if ( editable.contains( startNode ) && editable.contains( endNode ) ) {
						selection.selectBookmarks( [ bookmark ] );
					} else {
						// If one of bookmarks is not in DOM, clean up leftovers.
						startNode.remove();
						endNode.remove();
					}

					editor.fire( 'unlockSnapshot' );
				}

				notification.hide();
				finalizeCreation();
			},

			errorCallback: function() {
				notification.hide();
				editor.widgets.destroy( instance, true );
				editor.showNotification( lang.embeddingFailed, 'info' );
			}
		} );

		function finalizeCreation() {
			editor.widgets.finalizeCreation( temp );
		}
	}

	CKEDITOR.plugins.autoEmbed = {
		/**
		 * Gets the definition of the widget that should be used to automatically embed the specified link.
		 *
		 * This method uses the value of the {@link CKEDITOR.config#autoEmbed_widget} option.
		 *
		 * @since 4.5
		 * @member CKEDITOR.plugins.autoEmbed
		 * @param {CKEDITOR.editor} editor
		 * @param {String} url The URL to be embedded.
		 * @returns {CKEDITOR.plugins.widget.definition/null} The definition of the widget to be used to embed the link.
		 */
		getWidgetDefinition: function( editor, url ) {
			var opt = editor.config.autoEmbed_widget || 'embed,embedSemantic',
				name,
				widgets = editor.widgets.registered;

			if ( typeof opt == 'string' ) {
				opt = opt.split( ',' );

				while ( ( name = opt.shift() ) ) {
					if ( widgets[ name ] ) {
						return widgets[ name ];
					}
				}
			} else if ( typeof opt == 'function' ) {
				return widgets[ opt( url ) ];
			}

			return null;
		}
	};

	/**
	 * Specifies the widget to use to automatically embed a link. The default value
	 * of this option defines that either the [Media Embed](ckeditor.com/addon/embed) or
	 * [Semantic Media Embed](ckeditor.com/addon/embedsemantic) widgets will be used, depending on which is enabled.
	 *
	 * The general behavior:
	 *
	 * * If a string (widget names separated by commas) is provided, then the first of the listed widgets which is registered
	 *   will be used. For example, if `'foo,bar,bom'` is set and widgets `'bar'` and `'bom'` are registered, then `'bar'`
	 *   will be used.
	 * * If a callback is specified, then it will be executed with the URL to be embedded and it should return the
	 *   name of the widget to be used. It allows to use different embed widgets for different URLs.
	 *
	 * Example:
	 *
	 *		// Defines that embedSemantic should be used (regardless of whether embed is defined).
	 *		config.autoEmbed_widget = 'embedSemantic';
	 *
	 * Using with custom embed widgets:
	 *
	 *		config.autoEmbed_widget = 'customEmbed';
	 *
	 * **Note:** Plugin names are always lower case, while widget names are not, so widget names do not have to equal plugin names.
	 * For example, there is the `embedsemantic` plugin and the `embedSemantic` widget.
	 *
	 * Read more in the [documentation](#!/guide/dev_media_embed-section-automatic-embedding-on-paste)
	 * and see the [SDK sample](http://sdk.ckeditor.com/samples/mediaembed.html).
	 *
	 * @since 4.5
	 * @cfg {String/Function} [autoEmbed_widget='embed,embedSemantic']
	 * @member CKEDITOR.config
	 */
} )();
