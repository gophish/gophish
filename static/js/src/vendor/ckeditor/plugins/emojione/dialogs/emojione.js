CKEDITOR.dialog.add( 'emojioneDialog', function( editor ) {
	var config = editor.config, columns = 10, i;
	var dialog;
	var onClick = function( evt ) {
		var target = evt.data.getTarget();
		var unicode = target.getAttribute('data-unicode');
		if (unicode) {
			editor.insertText(emojione.convert(target.getAttribute('data-unicode')));
			dialog.hide();
		}
		evt.data.preventDefault();
	};

	var onKeydown = CKEDITOR.tools.addFunction( function( ev, element ) {
		ev = new CKEDITOR.dom.event( ev );
		element = new CKEDITOR.dom.element( element );
		var relative, nodeToMove;

		var keystroke = ev.getKeystroke(),
			rtl = editor.lang.dir == 'rtl';
		switch ( keystroke ) {
			// UP-ARROW
			case 38:
				// relative is TR
				if ( ( relative = element.getParent().getParent().getPrevious() ) ) {
					nodeToMove = relative.getChild( [ element.getParent().getIndex(), 0 ] );
					nodeToMove.focus();
				}
				ev.preventDefault();
				break;
			// DOWN-ARROW
			case 40:
				// relative is TR
				if ( ( relative = element.getParent().getParent().getNext() ) ) {
					nodeToMove = relative.getChild( [ element.getParent().getIndex(), 0 ] );
					if ( nodeToMove )
						nodeToMove.focus();
				}
				ev.preventDefault();
				break;
			// ENTER
			// SPACE
			case 32:
				onClick( { data: ev } );
				ev.preventDefault();
				break;

			// RIGHT-ARROW
			case rtl ? 37 : 39:
				// relative is TD
				if ( ( relative = element.getParent().getNext() ) ) {
					nodeToMove = relative.getChild( 0 );
					nodeToMove.focus();
					ev.preventDefault( true );
				}
				// relative is TR
				else if ( ( relative = element.getParent().getParent().getNext() ) ) {
					nodeToMove = relative.getChild( [ 0, 0 ] );
					if ( nodeToMove )
						nodeToMove.focus();
					ev.preventDefault( true );
				}
				break;

			// LEFT-ARROW
			case rtl ? 39 : 37:
				// relative is TD
				if ( ( relative = element.getParent().getPrevious() ) ) {
					nodeToMove = relative.getChild( 0 );
					nodeToMove.focus();
					ev.preventDefault( true );
				}
				// relative is TR
				else if ( ( relative = element.getParent().getParent().getPrevious() ) ) {
					nodeToMove = relative.getLast().getChild( 0 );
					nodeToMove.focus();
					ev.preventDefault( true );
				}
				break;
			default:
				// Do not stop not handled events.
				return;
		}
	} );

	var buildHtml = function(group) {
		var labelId = CKEDITOR.tools.getNextId() + '_smiley_emtions_label';
		var html = [
			'<div style="max-height:300px;overflow-y:scroll;">' +
			'<span id="' + labelId + '" class="cke_voice_label">Test</span>',
			'<table role="listbox" aria-labelledby="' + labelId + '" style="width:100%;height:100%;border-collapse:separate;" cellspacing="2" cellpadding="2"',
			CKEDITOR.env.ie && CKEDITOR.env.quirks ? ' style="position:absolute;"' : '',
			'><tbody>'
		];

		var list = {};
		var i = 0;
		emojione.imageType = 'svg'; // or svg
		emojione.sprites = true;
		emojione.imagePathSVGSprites = '/vendor/emojione/assets/sprites/emojione.sprites.svg';

		for (var shortcode in emojione.emojioneList) {

			if (!emojione.emojioneList.hasOwnProperty(shortcode)) continue;
			var obj = emojione.emojioneList[shortcode];
			if (!obj.isCanonical) continue;
			for (var prop in obj) {
				if(!obj.hasOwnProperty(prop)) continue;
				if (config.emojione.emojis[group].indexOf(shortcode) != -1) {
					list[shortcode] = obj;
				}
			}
		}

		for (var shortcode in list) {

			if ( i % columns === 0 )
				html.push( '<tr role="presentation">' );

			if (!list.hasOwnProperty(shortcode)) continue;

			var obj = list[shortcode];
			for (var prop in obj) {
				if(!obj.hasOwnProperty(prop)) continue;
			}

			html.push(
				'<td class="cke_centered" style="vertical-align: middle;" role="presentation">' +
				'<a style="font-size: 25px;" data-unicode="' + obj.unicode[0] + '" data-shortcode="' + shortcode + '" href="javascript:void(0)" role="option"', ' aria-posinset="' + ( i + 1 ) + '"', ' aria-setsize=""', ' aria-labelledby=""',
				' class="cke_hand" tabindex="-1" onkeydown="CKEDITOR.tools.callFunction( ', onKeydown, ', event, this );">',
				emojione.shortnameToUnicode(shortcode) +
				'</a>', '</td>'
			);

			if ( i % columns == columns - 1 )
				html.push( '</tr>' );
			i++;
		}


		if ( i < columns - 1 ) {
			for ( ; i < columns - 1; i++ )
				html.push( '<td></td>' );
			html.push( '</tr>' );
		}

		html.push( '</tbody></table></div>' );
		return html;
	};



	var emojis = function(group) {
		return {
			type: 'html',
			id: 'emojiSelector',
			html: buildHtml(group).join( '' ),
			onLoad: function( event ) {
				dialog = event.sender;
			},
			focus: function() {
				var self = this;
				setTimeout( function() {
					var firstSmile = self.getElement().getElementsByTag( 'a' ).getItem( 0 );
					firstSmile.focus();
				}, 0 );
			},
			onClick: onClick,
			style: 'width: 100%; border-collapse: separate;'
		}
	};

	return {
		title: 'Emojis',
		minWidth: 550,
		minHeight: 200,
		contents: [
			{
				id: 'tab-people',
				label: editor.config.emojione.tabs.people,
				elements: [
					emojis('people')
				]
			}, {
				id: 'tab-nature',
				label: editor.config.emojione.tabs.nature,
				elements: [
					emojis('nature')
				]
			}, {
				id: 'tab-objects',
				label: editor.config.emojione.tabs.objects,
				elements: [
					emojis('objects')
				]
			}, {
				id: 'tab-places',
				label: editor.config.emojione.tabs.places,
				elements: [
					emojis('places')
				]
			}, {
				id: 'tab-symbols',
				label: editor.config.emojione.tabs.symbols,
				elements: [
					emojis('symbols')
				]
			}
		],
		onShow: function() {

		}
	};
});