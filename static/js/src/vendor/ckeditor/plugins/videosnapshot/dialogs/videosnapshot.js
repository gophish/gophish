/*

	This file is a part of simplebuttion project.

	Copyright (C) Thanh D. Dang <thanhdd.it@gmail.com>

	simplebuttion is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	simplebuttion is distributed in the hope that it will be useful, but
	WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
	General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.

*/


CKEDITOR.dialog.add( 'videosnapshotDialog', function( editor ) {
	return {
		title: 'Video Snapshot',
		minWidth: 400,
		minHeight: 200,
		contents: [
			{
				id: 'tab-basic',
				elements: [
					{
						type: 'text',
						id: 'video-url',
						label: 'URL (Youtube)',
						validate: CKEDITOR.dialog.validate.notEmpty( "Text field cannot be empty." ),
						setup: function( element ) {
							this.setValue( element.getAttribute('href') );
						},
						commit: function( element ) {
							var play_image = CKEDITOR.plugins.getPath('videosnapshot') + 'images/play.png';
							element.setAttribute('href', this.getValue());
							var key = this.getValue().split('watch?v=')[1];
							if (key) {
								var video_url = 'https://www.youtube.com/watch?v=' + key;
								var image_url = 'https://i.ytimg.com/vi/' + key + '/hqdefault.jpg';
								element.setHtml('<img src="'+image_url+'" style="max-width:480px;margin:auto;"/><span style="background-image: url(' + play_image + '); background-repeat:no-repeat; background-position:center center; top:0; bottom:0; left:0; right:0; position:absolute; pointer-events:none; background-color:rgba(0,0,0,0.5);"> </span>');
							}

						}
					}
				]
			}
		],

		onShow: function() {
			var selection = editor.getSelection();
			var element = selection.getStartElement();
			if ( element && (element.hasClass('video-snapshot-plugin') || element.getParent().hasClass('video-snapshot-plugin')) ) {
				if (element.getParent().hasClass('video-snapshot-plugin'))
					element = element.getParent();

				this.insertMode = false;
			} else {
				element = editor.document.createElement( 'a' );
				element.setAttribute('class', 'video-snapshot-plugin');
				element.setAttribute('target', '_blank');
				element.setAttribute('style', 'display:block;max-width:480px;position:relative;margin:auto;');
				element.setAttribute('href', '');
				this.insertMode = true;
			}
			this.element = element;
			if (!this.insertMode)
				this.setupContent( this.element );
		},

		onOk: function() {
			var dialog = this;
			var video_snapshot = this.element;
			this.commitContent( video_snapshot );

			if ( this.insertMode )
				editor.insertElement( video_snapshot );
		}
	};
});
