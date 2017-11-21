/*

	This file is a part of videosnapshot project.

	Copyright (C) Thanh D. Dang <thanhdd.it@gmail.com>

	videosnapshot is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	videosnapshot is distributed in the hope that it will be useful, but
	WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
	General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.

*/


CKEDITOR.plugins.add( 'videosnapshot', {
	init: function( editor ) {
		editor.addCommand( 'videosnapshot', new CKEDITOR.dialogCommand( 'videosnapshotDialog' ) );
		editor.ui.addButton( 'videosnapshot', {
			label: 'Video Snapshot',
			command: 'videosnapshot',
			icon: this.path + 'images/videosnapshot.png'
		});
		CKEDITOR.dialog.add( 'videosnapshotDialog', this.path + 'dialogs/videosnapshot.js' );
	}
});
