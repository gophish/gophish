/**
 * @license Copyright (c) 2003-2017, CKSource - Frederico Knabben. All rights reserved.
 * For licensing, see LICENSE.md or http://ckeditor.com/license
 */

CKEDITOR.editorConfig = function( config ) {
	// Define changes to default configuration here. For example:
	// config.language = 'fr';
	// config.uiColor = '#AADC6E';
  // config.language_list = [ 'he:Hebrew:rtl', 'pt:Portuguese', 'de:German', 'ar:Arabic:rtl', 'fr:French', 'es:Spanish' ];
  config.extraPlugins = 'base64image,page2images,imagepaste,imagerotate,bgimage,autoembed,autolink,autosave,lineutils,widget,widgetselection,uploadwidget,filetools,notification,notificationaggregator,codesnippet,symbol,youtube,btgrid,emojione,chart,oembed,zoom,quicktable,texttransform,slideshow,cssanim,language,videosnapshot,dialogadvtab';
};
