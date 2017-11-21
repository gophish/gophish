/**
 * The slideshow dialog definition.
 * Copyright (c) 2003-2013, Cricri042. All rights reserved.
 * Targeted for "ad-gallery" JavaScript : http://adgallery.codeplex.com/
 * And "Fancybox" : http://fancyapps.com/fancybox/
 */
/**
 * Debug : var_dump
 *
 * @var: Var
 * @level: Level max
 *
 */

function removeDomainFromUrl(string) {
    "use strict";
    return string.replace(/^https?:\/\/[^\/]+/i, '');
}

var IMG_PARAM = {URL:0, TITLE:1, ALT:2, WIDTH:3, HEIGHT:4},
pluginPath = removeDomainFromUrl(CKEDITOR.plugins.get( 'slideshow' ).path),
BASE_PATH = removeDomainFromUrl(CKEDITOR.basePath),
//SCRIPT_JQUERY = "https://ajax.googleapis.com/ajax/libs/jquery/1.9.0/jquery.min.js",
SCRIPT_JQUERY =  pluginPath+"3rdParty/jquery.min.js",
SCRIPT_ADDGAL =  pluginPath+"3rdParty/ad-gallery/jquery.ad-gallery.min.js",
CSS_ADDGAL = pluginPath+"3rdParty/ad-gallery/jquery.ad-gallery.css",
SCRIPT_FANCYBOX = pluginPath+'3rdParty/fancybox2/jquery.fancybox.pack.js?v=2.1.5',
CSS_FANCYBOX = pluginPath+"3rdParty/fancybox2/jquery.fancybox.css?v=2.1.5";

function var_dump(_var, _level) {
  "use strict";
  var dumped_text = "";
  if(!_level) {
      _level = 0;
  }

  //The padding given at the beginning of the line.
  var level_padding = "";
  var j;
  for(j=0; j<_level+1; j+=1) {
      level_padding += "    ";
  }

    if(typeof(_var) == 'object') { //Array/Hashes/Objects
        var item;
        var value;

        for(item in _var) {
            if (_var.hasOwnProperty(item)) {
                value = _var[item];

                if(typeof(value) == 'object') { // If it is an array,
                  dumped_text += level_padding + "'" + item + "' ...\n";
                  dumped_text += var_dump(value, _level+1);
                } else {
                  dumped_text += level_padding +"'"+ item +"' => \""+ value +"\"\n";
                }
            }
        }
        
    } else { //Stings/Chars/Numbers etc.
        dumped_text = "===>"+ _var +"<===("+ typeof(_var) +")";
    }
  return dumped_text;
}

var listItem = function( node ) {
    "use strict";
    return node.type == CKEDITOR.NODE_ELEMENT && node.is( 'li' );
};

var ULItem = function( node ) {
    "use strict";
    return node.type == CKEDITOR.NODE_ELEMENT && node.is( 'ul' );
};

var iFrameItem = function( node ) {
    "use strict";
    return node.type == CKEDITOR.NODE_ELEMENT && node.is( 'iframe' );
};

Array.prototype.pushUnique = function (item){
    "use strict";
    var i;
    for ( i = 0; i < this.length ;  i+=1 ) {
        if (this[i][0] == item[0]) {
            return -1;
        }
    }
    this.push(item);
    return this.length - 1;
};

Array.prototype.updateVal = function (item, data){
    "use strict";
    var i;
    for ( i = 0; i < this.length ;  i+=1 ) {
            if (this[i][0] == item) {
                    this[i] = [item, data];
                    return true;
            }
    }
    this[i] = [item, data];
    return false;
};

Array.prototype.getVal = function (item){
    "use strict";
    var i;
    for ( i = 0; i < this.length ;  i+=1 ) {
            if (this[i][0] == item) {
                    return this[i][1];
            }
    }
    return null;
};


// Our dialog definition.
CKEDITOR.dialog.add( 'slideshowDialog', function( editor ) {
    "use strict";
    var lang = editor.lang.slideshow;

//----------------------------------------------------------------------------------------------------
// COMBO STUFF
//----------------------------------------------------------------------------------------------------
	// Add a new option to a CHKBOX object (combo or list).
	function addOption( combo, optionText, optionValue, documentObject, index )
	{
		combo = getSelect( combo );
		var oOption;
		if ( documentObject ) {
                    oOption = documentObject.createElement( "OPTION" );
                } else {
                    oOption = document.createElement( "OPTION" );
                }

		if ( combo && oOption && oOption.getName() == 'option' )
		{
			if ( CKEDITOR.env.ie ) {
				if ( !isNaN( parseInt( index, 10) ) ) {
					combo.$.options.add( oOption.$, index );
                                } else {
					combo.$.options.add( oOption.$ );
                                }

				oOption.$.innerHTML = optionText.length > 0 ? optionText : '';
				oOption.$.value     = optionValue;
			} else {
				if ( index !== null && index < combo.getChildCount() ) {
                                    combo.getChild( index < 0 ? 0 : index ).insertBeforeMe( oOption );
                                } else {
                                    combo.append( oOption );
                                }

				oOption.setText( optionText.length > 0 ? optionText : '' );
				oOption.setValue( optionValue );
			}
		} else {
			return false;
		}
		return oOption;
	}
        
	// Remove all selected options from a CHKBOX object.
	function removeSelectedOptions( combo )
	{
		combo = getSelect( combo );
		// Save the selected index
		var iSelectedIndex = getSelectedIndex( combo );
		// Remove all selected options.
                var i;
		for ( i = combo.getChildren().count() - 1 ; i >= 0 ; i-=1 )
		{
			if ( combo.getChild( i ).$.selected ) {
                            combo.getChild( i ).remove();
                        }
		}

		// Reset the selection based on the original selected index.
		setSelectedIndex( combo, iSelectedIndex );
	}
        
	//Modify option  from a CHKBOX object.
	function modifyOption( combo, index, title, value )
	{
		combo = getSelect( combo );
		if ( index < 0 ) {
                    return false;
                }
		var child = combo.getChild( index );
		child.setText( title );
		child.setValue( value );
		return child;
	}
        
	function removeAllOptions( combo )
	{
		combo = getSelect( combo );
		while ( combo.getChild( 0 ) && combo.getChild( 0 ).remove() )
		{ /*jsl:pass*/ }
	}
        
	// Moves the selected option by a number of steps (also negative).
	function changeOptionPosition( combo, steps, documentObject, dialog )
	{
		combo = getSelect( combo );
		var iActualIndex = getSelectedIndex( combo );
		if ( iActualIndex < 0 ) {
                    return false;
                }

		var iFinalIndex = iActualIndex + steps;
		iFinalIndex = ( iFinalIndex < 0 ) ? 0 : iFinalIndex;
		iFinalIndex = ( iFinalIndex >= combo.getChildCount() ) ? combo.getChildCount() - 1 : iFinalIndex;

		if ( iActualIndex == iFinalIndex ) {
                    return false;
                }

		var re = /(^IMG_\d+)/;
		// Modify sText in final index
		var oOption = combo.getChild( iFinalIndex ),
		sText	= oOption.getText(),
		sValue	= oOption.getValue();
		sText = sText.replace(re, "IMG_"+iActualIndex);
		modifyOption( combo, iFinalIndex, sText, sValue );

		// do the move
		oOption = combo.getChild( iActualIndex );
                sText	= oOption.getText();
                sValue	= oOption.getValue();

		oOption.remove();

//		alert(sText+ " / "+ sValue);
//		var result = re.exec(sText);
		sText = sText.replace(re, "IMG_"+iFinalIndex);
//		alert(sText);
		oOption = addOption( combo, sText, sValue, ( !documentObject ) ? null : documentObject, iFinalIndex );
		setSelectedIndex( combo, iFinalIndex );

		// update dialog.imagesList
		var valueActual = dialog.imagesList[iActualIndex];
		var valueFinal = dialog.imagesList[iFinalIndex];
		dialog.imagesList[iActualIndex] = valueFinal;
		dialog.imagesList[iFinalIndex] = valueActual;

		return oOption;
	}
        
	function getSelectedIndex( combo )
	{
		combo = getSelect( combo );
		return combo ? combo.$.selectedIndex : -1;
	}
        
	function setSelectedIndex( combo, index )
	{
		combo = getSelect( combo );
		if ( index < 0 ) {
                    return null;
                }

                var count = combo.getChildren().count();
		combo.$.selectedIndex = ( index >= count ) ? ( count - 1 ) : index;
		return combo;
	}
        
	function getOptions( combo )
	{
		combo = getSelect( combo );
		return combo ? combo.getChildren() : false;
	}
        
	function getSelect( obj )
	{
		if ( obj && obj.domId && obj.getInputElement().$ ) {
                    return  obj.getInputElement();
                } else if ( obj && obj.$ ) {
                    return obj;
                }
		return false;
	}

	function unselectAll(dialog) {
		var editBtn = dialog.getContentElement( 'slideshowinfoid', 'editselectedbtn');
		var deleteBtn = dialog.getContentElement( 'slideshowinfoid', 'removeselectedbtn');
		editBtn = getSelect( editBtn );
		editBtn.hide();
		deleteBtn = getSelect( deleteBtn );
		deleteBtn.hide();
		var comboList = dialog.getContentElement( 'slideshowinfoid', 'imglistitemsid');
		comboList = getSelect( comboList );
                var i;
		for ( i = comboList.getChildren().count() - 1 ; i >= 0 ; i-=1 )
		{
			comboList.getChild( i ).$.selected = false;
		}
	}

	function unselectIfNotUnique(combo) {
		var dialog = combo.getDialog();
		var selectefItem = null;
		combo = getSelect( combo );
		var cnt = 0;
		var editBtn = dialog.getContentElement( 'slideshowinfoid', 'editselectedbtn');
		var deleteBtn = dialog.getContentElement( 'slideshowinfoid', 'removeselectedbtn');
                var i, child;
		for ( i = combo.getChildren().count() - 1 ; i >= 0 ; i-=1 )
		{
			child = combo.getChild( i );
			if ( child.$.selected ) {
				cnt++;
				selectefItem = child;
			}
		}
		if (cnt > 1) {
			unselectAll(dialog);
			return null;
		} else if (cnt == 1) {
				editBtn = getSelect( editBtn );
				editBtn.show();
				deleteBtn = getSelect( deleteBtn );
				deleteBtn.show();
				displaySelected(dialog);
				return selectefItem;
		}
		return null;
	}

	function displaySelected (dialog) {
		if (dialog.openCloseStep == true) {
                    return;
                }
		var previewCombo = dialog.getContentElement( 'slideshowinfoid', 'framepreviewid');
		if ( previewCombo.isVisible()) {
			previewSlideShow(dialog);
		} else {
			editeSelected(dialog);
		}
	}

	function selectFirstIfNotUnique(combo) {
		var dialog = combo.getDialog();
		combo = getSelect( combo );
		var firstSelectedInd = 0;
                var i, child, selectefItem;
		for ( i = 0; i < combo.getChildren().count()  ; i+=1 )
		{
			child = combo.getChild( i );
			if ( child.$.selected ) {
				selectefItem = child;
				firstSelectedInd = i;
				break;
			}
		}
		setSelectedIndex(combo, firstSelectedInd);
		displaySelected(dialog);
	}

	function getSlectedIndex(dialog) {
		var combo = dialog.getContentElement( 'slideshowinfoid', 'imglistitemsid');
		return getSelectedIndex( combo );
	}

//----------------------------------------------------------------------------------------------------
//----------------------------------------------------------------------------------------------------

	function removePlaceHolderImg(dialog) {
		var urlPlaceHolder =  BASE_PATH  + 'plugins/slideshow/images/placeholder.png' ;
		if ((dialog.imagesList.length == 1) && (dialog.imagesList[0][IMG_PARAM.URL] == urlPlaceHolder)) {
			// Remove the place Holder Image
			var combo = dialog.getContentElement( 'slideshowinfoid', 'imglistitemsid');
			combo = getSelect( combo );
			var i = 0;
			// Remove image from image Array
			dialog.imagesList.splice(i, 1);
			// Remove image from combo image list
			combo.getChild( i ).remove();
		}
	}

	function updateImgList(dialog) {
		removePlaceHolderImg(dialog);
		var preview = dialog.previewImage;
		var url = preview.$.src;
		var ratio = preview.$.width / preview.$.height;
		var w = 50;
		var h = 50;
		if (ratio > 1) {
			h = h/ratio;
		} else {
			w = w*ratio;
		}
                var oOption;
                var combo = dialog.getContentElement( 'slideshowinfoid', 'imglistitemsid');
		var ind = dialog.imagesList.pushUnique([url, '', '', w.toFixed(0), h.toFixed(0)]);
		if (ind >= 0) {
			oOption = addOption( combo, 'IMG_'+ind + ' : ' + url.substring(url.lastIndexOf('/')+1), url, dialog.getParentEditor().document );
			// select index 0
			setSelectedIndex(combo, ind);
			// Update dialog
			displaySelected(dialog);
		}
	}

	function editeSelected(dialog) {
		var combo = dialog.getContentElement( 'slideshowinfoid', 'imglistitemsid');
		var iSelectedIndex = getSelectedIndex( combo );
		var value = dialog.imagesList[iSelectedIndex];

		combo = dialog.getContentElement( 'slideshowinfoid', 'imgtitleid');
		combo = getSelect( combo );
		combo.setValue(value[1]);
		combo = dialog.getContentElement( 'slideshowinfoid', 'imgdescid');
		combo = getSelect( combo );
		combo.setValue(value[2]);
		combo = dialog.getContentElement( 'slideshowinfoid', 'imgpreviewid');
		combo = getSelect( combo );
		//console.log( "VALUE IMG -> " +  value[iSelectedIndex] );
		var imgHtml = '<div style="text-align:center;"> <img src="'+ value[0] +
						'" title="' + value[1] +
						'" alt="' + value[2] +
						'" style=" max-height: 200px;  max-width: 350px;' + '"> </div>';
		combo.setHtml(imgHtml);
		var previewCombo = dialog.getContentElement( 'slideshowinfoid', 'framepreviewid');
		var imgCombo =  dialog.getContentElement( 'slideshowinfoid', 'imgparamsid');
		previewCombo = getSelect( previewCombo );
		previewCombo.hide();
		imgCombo = getSelect( imgCombo );
		imgCombo.show();
	}

	function removeSelected(dialog) {
		var combo = dialog.getContentElement( 'slideshowinfoid', 'imglistitemsid');
		combo = getSelect( combo );
		var someRemoved = false;
		// Remove all selected options.
                var i;
		for ( i = combo.getChildren().count() - 1 ; i >= 0 ; i-- )
		{
			if ( combo.getChild( i ).$.selected ) {
				// Remove image from image Array
				dialog.imagesList.splice(i, 1);
				// Remove image from combo image list
				combo.getChild( i ).remove();
				someRemoved = true;
			}
		}
		if (someRemoved) {
			if (dialog.imagesList.length == 0) {
				var url =  BASE_PATH  + 'plugins/slideshow/images/placeholder.png' ;
				var oOption = addOption( combo, 'IMG_0' + ' : ' + url.substring(url.lastIndexOf('/')+1) , url, dialog.getParentEditor().document );
				 dialog.imagesList.pushUnique([url, lang.imgTitle, lang.imgDesc, '50', '50']);
			}
			// select index 0
			setSelectedIndex(combo, 0);
			// Update dialog
			displaySelected(dialog);
		}
	}

	function upDownSelected(dialog, offset) {
		var combo = dialog.getContentElement( 'slideshowinfoid', 'imglistitemsid');
		combo = getSelect( combo );
		var iSelectedIndex = getSelectedIndex( combo );
		if (combo.getChildren().count() == 1) {
                    return;
                }
		if ((offset == -1) && (iSelectedIndex == 0)) {
                    return;
                }
		if ((offset == 1) && (iSelectedIndex == combo.getChildren().count()-1)) {
                    return;
                }
		//alert(iSelectedIndex+" / "+combo.getChildren().count() + " / "+ offset);
		changeOptionPosition( combo, offset, dialog.getParentEditor().document, dialog );

		updateFramePreview(dialog);
	}
        
	// To automatically get the dimensions of the poster image
	var onImgLoadEvent = function() {
		// Image is ready.
		var preview = this.previewImage;
		preview.removeListener( 'load', onImgLoadEvent );
		preview.removeListener( 'error', onImgLoadErrorEvent );
		preview.removeListener( 'abort', onImgLoadErrorEvent );
		//console.log( "previewImage -> " + preview );
		updateImgList(this);
	};

	var onImgLoadErrorEvent = function() {
		// Error. Image is not loaded.
		var preview = this.previewImage;
		preview.removeListener( 'load', onImgLoadEvent );
		preview.removeListener( 'error', onImgLoadErrorEvent );
		preview.removeListener( 'abort', onImgLoadErrorEvent );
	};

	function updateTitle(dialog, val) {
		dialog.imagesList[getSlectedIndex(dialog)][IMG_PARAM.TITLE] = val;
		editeSelected(dialog);
	}

	function updateDescription(dialog, val) {
		dialog.imagesList[getSlectedIndex(dialog)][IMG_PARAM.ALT] = val;
		editeSelected(dialog);
	}

	function previewSlideShow(dialog) {
		var previewCombo = dialog.getContentElement( 'slideshowinfoid', 'framepreviewid');
		var imgCombo =  dialog.getContentElement( 'slideshowinfoid', 'imgparamsid');
		imgCombo = getSelect( imgCombo );
		imgCombo.hide();
		previewCombo = getSelect( previewCombo );
		previewCombo.show();
		updateFramePreview(dialog);
	}

	function feedFrame(frame, data) {
		frame.open();
		frame.writeln( data );
		frame.close();
	}

// 	function unprotectRealComments( html )
// 	{
// 		return html.replace( /<!--\{cke_protected\}\{C\}([\s\S]+?)-->/g,
// 			function( match, data )
// 			{
// 				return decodeURIComponent( data );
// 			});
// 	};
//
// 	function unprotectSource( html, editor )
// 	{
// 		return html.replace( /<!--\{cke_protected\}([\s\S]+?)-->/g, function( match, data )
// 			{
// 				return decodeURIComponent( data );
// 			});
// 	}

	function updateFramePreview(dialog) {
		var width = 436;
		var height = 300;
		if ( dialog.params.getVal('showthumbid') == true) {
			height -= 120;
		} else if ( dialog.params.getVal('showcontrolid') == true) {
			height -= 30;
		}
		if (dialog.imagesList.length == 0) {
                    return;
                }
		var combo = dialog.getContentElement( 'slideshowinfoid', 'imglistitemsid');
		var iSelectedIndex = getSelectedIndex( combo );
		if (iSelectedIndex < 0) {
                    iSelectedIndex = 0;
                }
                
		combo = dialog.getContentElement( 'slideshowinfoid', 'framepreviewid');
                
		var strVar="";
		var jqueryStr = '<script src="'+SCRIPT_JQUERY+'" type="text/javascript"></script>';
		strVar += "<head>";
//		if (editor.config.slideshowDoNotLoadJquery && (editor.config.slideshowDoNotLoadJquery == true)) {
//			jqueryStr = '';
//		}
		strVar += jqueryStr;
        strVar += "<script type=\"text\/javascript\" src=\""+SCRIPT_ADDGAL+"\"><\/script>";
        strVar += "<link rel=\"stylesheet\" type=\"text\/css\" href=\""+CSS_ADDGAL+"\" \/>";
		if ( dialog.params.getVal('openOnClickId') == true) {
		    strVar += "<link rel=\"stylesheet\" type=\"text\/css\" href=\""+CSS_FANCYBOX+"\" \/>";
		    strVar += "<script type=\"text\/javascript\" src=\""+SCRIPT_FANCYBOX+"\"><\/script>";
		    strVar += "<script type=\"text\/javascript\">";
		    strVar += 	createScriptFancyBoxRun(dialog);
		    strVar += "<\/script>";
		}

	    strVar += "<script type=\"text\/javascript\">";
	    strVar += 	createScriptAdGalleryRun(dialog, iSelectedIndex, width, height);
	    strVar += "<\/script>";

	    strVar += "<\/head>";
	    strVar += "<body>";
	    var domGallery = createDOMdGalleryRun(dialog);
            strVar += domGallery.getOuterHtml();
	    strVar += "<\/body>";
	    strVar += "";

            combo = getSelect( combo );
            var theFrame = combo.getFirst( iFrameItem );
            if (theFrame) {
                theFrame.remove();
            }
	    var ifr = null;

	    var w = width+60;
	    var h = height;
		
            if ( dialog.params.getVal('showthumbid') == true) {
                    h += 120;
            } else if ( dialog.params.getVal('showcontrolid') == true) {
                    h += 30;
            }
            var iframe = CKEDITOR.dom.element.createFromHtml( '<iframe' +
                            ' style="width:'+w+'px;height:'+h+'px;background:azure; "'+
                            ' class="cke_pasteframe"' +
                            ' frameborder="10" ' +
                            ' allowTransparency="false"' +
//				' src="' + 'data:text/html;charset=utf-8,' +  strVar + '"' +
                            ' role="region"' +
                            ' scrolling="no"' +
                            '></iframe>' );

            iframe.setAttribute('name', 'totoFrame');
            iframe.setAttribute('id', 'totoFrame');
            iframe.on( 'load', function( event ) {
                    if (ifr != null) {
                        return;
                    }
                    ifr =  this.$;
                    var iframedoc;
                    if (ifr.contentDocument) {
                            iframedoc = ifr.contentDocument;
                    } else if (ifr.contentWindow) {
                            iframedoc = ifr.contentWindow.document;
                    }
                    
                    if (iframedoc){
                             // Put the content in the iframe
                             feedFrame(iframedoc, strVar);
                    } else {
                           //just in case of browsers that don't support the above 3 properties.
                           //fortunately we don't come across such case so far.
                           alert('Cannot inject dynamic contents into iframe.');
                    }
            });
            combo.append(iframe);
	}

	function initImgListFromDOM(dialog, slideShowContainer) {
		var i, image, src;
		var imgW, imgH;
                var ratio, w, h, ind;
		var arr  = slideShowContainer.$.getElementsByTagName("img");
		var combo = dialog.getContentElement( 'slideshowinfoid', 'imglistitemsid');
		for (i = 0; i < arr.length; i+=1) {
			image = arr[i];
			src = image.src;
			// IE Seems sometime to return 0 !!, So natural Width and Height seems OK
			// If not just pput 50, Not as good but not so bad !!
			imgW =  image.width;
			if (imgW == 0) {
                            imgW = image.naturalWidth;
                        }
			if (imgW == 0) {
				imgW = 50;
				imgH = 50;
			} else {
				imgH =  image.height;
				if (imgH == 0) {
                                    imgH = image.naturalHeight;
                                }
				if (imgH == 0) {
					imgW = 50;
					imgH = 50;
				}
			}
			ratio = imgW / imgH;
			w = 50;
			h = 50;
			if (ratio > 1) {
				h = h/ratio;
			} else {
				w = w*ratio;
			}
			ind = dialog.imagesList.pushUnique([src, image.title, image.alt, w, h]);
                        var oOption;
			if (ind >= 0) {
				oOption = addOption( combo, 'IMG_'+ind + ' : ' + src.substring(src.lastIndexOf('/')+1), src, dialog.getParentEditor().document );
			}
		}
		// select index 0
		setSelectedIndex(combo, 0);
		// Update dialog
		displaySelected(dialog);
        }

	function initImgListFromFresh(dialog) {
		var combo = dialog.getContentElement( 'slideshowinfoid', 'imglistitemsid');
		var url =  BASE_PATH  + 'plugins/slideshow/images/placeholder.png' ;
		var oOption = addOption( combo, 'IMG_0' + ' : ' + url.substring(url.lastIndexOf('/')+1) , url, dialog.getParentEditor().document );
		dialog.imagesList.pushUnique([url, lang.imgTitle, lang.imgDesc, '50', '50']);
		// select index 0
		setSelectedIndex(combo, 0);
		// Update dialog
		displaySelected(dialog);
        }


	function commitSlideShow(dialog) {
		dialog.slideshowDOM.setAttribute('data-'+this.id, this.getValue());
	}

	function loadValue() {
		var dialog = this.getDialog();
		if (dialog.newSlideShowMode) {
			// New fresh SlideShow so let's put dom data attributes from dialog default values
			dialog.slideshowDOM.setAttribute('data-'+this.id, this.getValue());
			switch ( this.type ) {
                            case 'checkbox':
                                    break;
                            case 'text':
                                    break;
                            case 'select':
                                    break;
                            default:
                                break;
			}
		} else {
			// Loaded SlideShow, so update Dialog values from DOM data attributes

			switch ( this.type ) {
                            case 'checkbox':
                                    this.setValue(dialog.slideshowDOM.getAttribute('data-'+this.id) == 'true');
                                    break;
                            case 'text':
                                    this.setValue(dialog.slideshowDOM.getAttribute('data-'+this.id));
                                    break;
                            case 'select':
                                    this.setValue(dialog.slideshowDOM.getAttribute('data-'+this.id));
                                    break;
                            default:
                                break;
			}
		}
	}

	function commitValue() {
		var dialog = this.getDialog();
		dialog.params.updateVal(this.id, this.getValue());
		switch ( this.type ) {
                    case 'checkbox':
                            break;
                    case 'text':
                            break;
                    case 'select':
                            break;
                    default:
                        break;
		}
		displaySelected(dialog);
	}

	function cleanAll(dialog) {
		if ( dialog.previewImage )
		{
			dialog.previewImage.removeListener( 'load', onImgLoadEvent );
			dialog.previewImage.removeListener( 'error', onImgLoadErrorEvent );
			dialog.previewImage.removeListener( 'abort', onImgLoadErrorEvent );
			dialog.previewImage.remove();
			dialog.previewImage = null;		// Dialog is closed.
		}
		dialog.imagesList = null;
		dialog.params = null;
		dialog.slideshowDOM = null;
		var combo = dialog.getContentElement( 'slideshowinfoid', 'imglistitemsid');
		removeAllOptions(combo);
		dialog.openCloseStep = false;

	}

	function randomChars(len) {
	    var chars = '';
	    while (chars.length < len) {
	        chars += Math.random().toString(36).substring(2);
	    }
	    // Remove unnecessary additional characters.
	    return chars.substring(0, len);
	}

	var numbering = function( id ) {
		//return CKEDITOR.tools.getNextId() + '_' + id;
		return 'cke_' + randomChars(8) + '_' + id;
	};

	function getImagesContainerBlock(dialog, dom) {
		var obj = dom.getElementsByTag("ul");
		if (obj == null) {
                    return null;
                }
		if (obj.count() == 1) {
			return obj.getItem(0);
		}
		return null;
	}

	function createScriptAdGalleryRun(dialog, iSelectedIndex, width, height) {
            var slideshowid =  dialog.params.getVal('slideshowid'),
                galleryId   =  'ad-gallery_' + slideshowid,
                strVar      = '(function($) {',
                strHook     = '';
        
	    strVar += "$(function() {";
//	    if (width == 0) width = "false";
	    if (height == 0) {
                height = dialog.params.getVal('pictheightid');
            }
//	    if (width == 0) width = dialog.params.getVal('pictWidthtid');
//	    if (height == 0) height = "false";
//	    if (width <= 1) width = "false";
	    if (width == 0) {
                width = "false";
            }
	    if (dialog.params.getVal('showtitleid') == false) {
	    	strHook = ",  hooks: { displayDescription: function(image) {}}";
	    }
	    var params = "loader_image: '"+pluginPath+"3rdParty/ad-gallery/loader.gif'," +
	    				" width:" + width + ", height:" + height +
	    				", start_at_index: " + iSelectedIndex +
	    				", animation_speed: " + dialog.params.getVal('animspeedid') + strHook +
	    				", update_window_hash: false, effect: '" + dialog.params.getVal('transitiontypeid') +
	    				"',";
	    //alert(params);

	    var slideShowParams = " slideshow: { enable: true, autostart: " + dialog.params.getVal('autostartid') +
											", start_label: '" + lang.labelStart + "'" +
											", stop_label: '" + lang.labelStop + "'" +
	    									", speed: " + dialog.params.getVal('speedid') * 1000 +
	    									"}";
	    strVar += "   var galleries = $('#"+galleryId+"').adGallery({" + params + slideShowParams + "});";
	    strVar += "});";
		strVar += "})(jQuery);";

            return strVar;
	}

//	function createScriptFancyBoxRun(dialog) {
//		var slideshowid =  dialog.params.getVal('slideshowid'),
//			galleryId   =  'ad-gallery_' + slideshowid,
//			str         = '(function($) {';
////		str +=  "$(document).ready(function() {";
//		str += "$(function() {";
//		str += "$(\"#"+galleryId+"\").on(\"click\",\".ad-image\",function(){";
//		str += "var imgObj =$(this).find(\"img\");";
//		str += "var isrc=imgObj.attr(\"src\");";
//		str += "var ititle=null;";
//		str += "var idesc=null;";
//		str += "var iname=isrc.split('/');";
//		str += "iname=iname[iname.length-1];";
//		str += "var imgdescid=$(this).find(\".ad-image-description\");";
//		str += "if(imgdescid){";
//		str += "ititle=$(this).find(\".ad-description-title\");";
//		str += "if(ititle)ititle=ititle.text();";
//		str += "if(ititle!='')ititle='<big>'+ititle+'</big>';";
//		str += "idesc=$(this).find(\"span\");";
//		str += "if(idesc)idesc=idesc.text();";
//		str += "if(idesc!=''){";
//		str += "if(ititle!='')ititle=ititle+'<br>';";
//		str += "idesc='<i>'+idesc+'</i>';";
//		str += "}";
//		str += "}";
//		str += "$.fancybox.open({";
//		str += "href:isrc,";
//		str += "beforeLoad:function(){";
//		str += "this.title=ititle+idesc;";
//		str += "},";
//		str += "});";
//		str += "});";
//		str += "});";
//		str += "})(jQuery);";
//
//                return str;
//	}
	function createScriptFancyBoxRun(dialog) {
		var slideshowid =  dialog.params.getVal('slideshowid'),
			galleryId   =  'ad-gallery_' + slideshowid,
			str         = '(function($) {';
		str += "$(function() {";
		str += "$(\"#"+galleryId+"\").on(\"click\",\".ad-image\",function(){";
		str += "var imgObj =$(this).find(\"img\");";
		str += "var isrc=imgObj.attr(\"src\");";
		str += "var ititle=null;";
		str += "var idesc=null;";
		str += "var iname=isrc.split('/');";
		str += "iname=iname[iname.length-1];";
		str += "var imgdescid=$(this).find(\".ad-image-description\");";
		str += "if(imgdescid){";
		str += "ititle=$(this).find(\".ad-description-title\");";
		str += "if(ititle)ititle=ititle.text();";
		str += "if(ititle!='')ititle='<big>'+ititle+'</big>';";
		str += "idesc=$(this).find(\"span\");";
		str += "if(idesc)idesc=idesc.text();";
//		str += 'console.log("idesc:", idesc);';
		str += "if (idesc.indexOf('IMAGE_LINK_') >= 0) {";
		str += "idesc = '';";
		str += "}";
		str += "if(idesc!=''){";
		str += "if(ititle!='')ititle=ititle+'<br>';";
		str += "idesc='<i>'+idesc+'</i>';";
		str += "}";
		str += "}";
		str += "$.fancybox.open({";
		str += "href:isrc,";
		str += "beforeLoad:function(){";
		str += "this.title=ititle+idesc;";
		str += "},";
		str += "});";
		str += "});";
		str += "});";
		str += "})(jQuery);";
        return str;
	}
	function createScriptLinkRun(dialog) {
		var slideshowid =  dialog.params.getVal('slideshowid'),
		galleryId   =  'ad-gallery_' + slideshowid,
		str         = '(function($) {';
		str += "$(function() {";
		str += "$(\"#"+galleryId+"\").on(\"click\",\".ad-image\",function(){";
		str += "var imgObj =$(this).find(\"img\");";
		str += "var isrc=imgObj.attr(\"src\");";
		str += "var ititle=null;";
		str += "var idesc=null;";
		str += "var iname=isrc.split('/');";
		str += "iname=iname[iname.length-1];";
		str += "var imgdescid=$(this).find(\".ad-image-description\");";
		str += "if(imgdescid){";
		str += "ititle=$(this).find(\".ad-description-title\");";
		str += "if(ititle)ititle=ititle.text();";
		str += "idesc=$(this).find(\"span\");";
	//	str += "console.log('desc0', idesc);";
		str += "if(idesc)idesc=idesc.text();";
		str += "if(idesc!=''){";
	//	str += "console.log('desc1', idesc);";
		
	//	str += "if (idesc.indexOf('LIEN:') == 0) {";
	//	str += "idesc = idesc.substring(5);}";
	//	str += "window.open(idesc);"
	//	str += "}";
		str += "var url=window.location.href.trim();";
		str += "if (idesc.indexOf('IMAGE_LINK_TAB:') >= 0) {";
		str += "	idesc = idesc.substring(15).trim();";
		str += " if (url != idesc) window.open(idesc,'_blank');";
		str += "} else if (idesc.indexOf('IMAGE_LINK_PAR:') >= 0) {";
		str += "	idesc = idesc.substring(15).trim();";
		str += " if (url != idesc) window.open(idesc,'_self');";
		str += "}";
	
		str += "}";
		str += "}";
		str += "});";
		str += "});";
		str += "})(jQuery);";
	    return str;
	}
	function feedUlWithImages(dialog, ulObj) {
                var i, liObj, aObj, newImgDOM;
		for ( i = 0; i < dialog.imagesList.length  ; i+=1 ) {
			liObj = ulObj.append( 'li' );
			liObj.setAttribute( 'contenteditable', 'false');
			aObj = liObj.append( 'a' );
			aObj.setAttribute( 'href', removeDomainFromUrl(dialog.imagesList[i][IMG_PARAM.URL]) );
			aObj.setAttribute('contenteditable', 'false');
			newImgDOM = aObj.append('img');
			newImgDOM.setAttribute( 'src', removeDomainFromUrl(dialog.imagesList[i][IMG_PARAM.URL]) );
			newImgDOM.setAttribute( 'title', dialog.imagesList[i][IMG_PARAM.TITLE]);
			newImgDOM.setAttribute( 'alt', dialog.imagesList[i][IMG_PARAM.ALT]);
			newImgDOM.setAttribute( 'contenteditable', 'false');
			newImgDOM.setAttribute('width',  dialog.imagesList[i][IMG_PARAM.WIDTH]);
			newImgDOM.setAttribute('height',  dialog.imagesList[i][IMG_PARAM.HEIGHT]);
		}
	}

	function createDOMdGalleryRun(dialog) {
		var slideshowid =  dialog.params.getVal('slideshowid');
		var galleryId =  'ad-gallery_' + slideshowid;
		var displayThumbs = 'display: block;';
		var displayControls = 'display: block;';
		if ( dialog.params.getVal('showthumbid') == false) {
			displayThumbs = 'display: none;';
		}
		if ( dialog.params.getVal('showcontrolid') == false) {
			displayControls = 'visibility: hidden;';
		}
		var slideshowDOM = editor.document.createElement( 'div' );
		slideshowDOM.setAttribute('id', slideshowid );
		slideshowDOM.setAttribute( 'class', 'slideshowPlugin');
		slideshowDOM.setAttribute( 'contenteditable', 'false');

		var galleryDiv =  slideshowDOM.append('div');
		galleryDiv.setAttribute( 'class', 'ad-gallery');
		galleryDiv.setAttribute( 'contenteditable', 'false');
		galleryDiv.setAttribute( 'id', galleryId);

		var wrapperObj =  galleryDiv.append('div');
		wrapperObj.setAttribute( 'class', 'ad-image-wrapper');
		wrapperObj.setAttribute( 'contenteditable', 'false');

		var controlObj =  galleryDiv.append('div');
		controlObj.setAttribute( 'class', 'ad-controls');
		controlObj.setAttribute( 'contenteditable', 'false');
		controlObj.setAttribute( 'style', displayControls);

		var navObj =  galleryDiv.append('div');
		navObj.setAttribute( 'class', 'ad-nav');
		navObj.setAttribute( 'style', displayThumbs);
		navObj.setAttribute( 'contenteditable', 'false');

		var thumbsObj =  navObj.append('div');
		thumbsObj.setAttribute( 'class', 'ad-thumbs');
		thumbsObj.setAttribute( 'contenteditable', 'false');

		var ulObj = thumbsObj.append('ul');
		ulObj.setAttribute('class', 'ad-thumb-list');
		ulObj.setAttribute( 'contenteditable', 'false');

		feedUlWithImages(dialog, ulObj);
		return slideshowDOM;
	}

	function ClickOkBtn(dialog) {
		var extraStyles = {},
		extraAttributes = {};

		dialog.openCloseStep = true;

		// Invoke the commit methods of all dialog elements, so the dialog.params array get Updated.
		dialog.commitContent(dialog);

		// Create a new DOM
                var slideshowDOM = createDOMdGalleryRun(dialog);

                // Add data tags to dom
                var i;
		for ( i = 0; i < dialog.params.length  ; i+=1 ) {
			slideshowDOM.data(dialog.params[i][0], dialog.params[i][1]);
		}
		if (!(editor.config.slideshowDoNotLoadJquery && (editor.config.slideshowDoNotLoadJquery == true))) {
	        var scriptjQuery =  CKEDITOR.document.createElement( 'script', {
				attributes: {
					type: 'text/javascript',
					src: SCRIPT_JQUERY
				}
			});
			slideshowDOM.append(scriptjQuery);
		}
		// Add javascript for ""ad-gallery"
		// Be sure the path is correct and file is available !!
		var scriptAdGallery =  CKEDITOR.document.createElement( 'script', {
			attributes: {
				type: 'text/javascript',
				src: SCRIPT_ADDGAL
			}
		});
		slideshowDOM.append(scriptAdGallery);

		if ( dialog.params.getVal('openOnClickId') == true) {
			// Dynamically add CSS for "fancyBox"
			// Be sure the path is correct and file is available !!
			var scriptFancyBoxCss =  CKEDITOR.document.createElement( 'script', {
				attributes: {
					type: 'text/javascript'
				}
			});
			scriptFancyBoxCss.setText("(function($) { $('head').append('<link rel=\"stylesheet\" href=\""+CSS_FANCYBOX+"\" type=\"text/css\" />'); })(jQuery);");
			slideshowDOM.append(scriptFancyBoxCss);

			// Add javascript for ""fancyBox"
			// Be sure the path is correct and file is available !!
			var scriptFancyBox =  CKEDITOR.document.createElement( 'script', {
				attributes: {
					type: 'text/javascript',
					src: SCRIPT_FANCYBOX
				}
			});
			slideshowDOM.append(scriptFancyBox);

			// Add RUN javascript for "fancybox"
			var scriptFancyboxRun =  CKEDITOR.document.createElement( 'script', {
				attributes: {
					type: 'text/javascript'
				}
			});
			scriptFancyboxRun.setText(createScriptFancyBoxRun(dialog));
			slideshowDOM.append(scriptFancyboxRun);
		}
		// Add RUN javascript for "link"
		var scriptLinkRun =  CKEDITOR.document.createElement( 'script', {
			attributes: {
				type: 'text/javascript'
			}
		});
		scriptLinkRun.setText(createScriptLinkRun(dialog));
		slideshowDOM.append(scriptLinkRun);

		// Dynamically add CSS for "ad-gallery"
		// Be sure the path is correct and file is available !!
		var scriptAdGalleryCss =  CKEDITOR.document.createElement( 'script', {
			attributes: {
				type: 'text/javascript'
			}
		});
		scriptAdGalleryCss.setText("(function($) { $('head').append('<link rel=\"stylesheet\" href=\""+CSS_ADDGAL+"\" type=\"text/css\" />'); })(jQuery);");
		slideshowDOM.append(scriptAdGalleryCss);

		// Add RUN javascript for "ad-Gallery"
		var scriptAdGalleryRun =  CKEDITOR.document.createElement( 'script', {
			attributes: {
				type: 'text/javascript'
			}
		});
		scriptAdGalleryRun.setText(createScriptAdGalleryRun(dialog, 0, 0, 0));
		slideshowDOM.append(scriptAdGalleryRun);

		if (dialog.imagesList.length) {
			extraStyles.backgroundImage =  'url("' + dialog.imagesList[0][IMG_PARAM.URL] + '")';
		}
		extraStyles.backgroundSize = '50%';
		extraStyles.display = 'block';
		// Create a new Fake Image
		var newFakeImage = editor.createFakeElement( slideshowDOM, 'cke_slideShow', 'slideShow', false );
		newFakeImage.setAttributes( extraAttributes );
		newFakeImage.setStyles( extraStyles );

		if ( dialog.fakeImage )
		{
			newFakeImage.replace( dialog.fakeImage );
			editor.getSelection().selectElement( newFakeImage );
		}
		else
		{
			editor.insertElement( newFakeImage );
		}

		cleanAll(dialog);
		dialog.hide();
		return true;
	}

	return {
		// Basic properties of the dialog window: title, minimum size.
		title : lang.dialogTitle,
		width: 500,
		height: 600,
		resizable: CKEDITOR.DIALOG_RESIZE_NONE,
		buttons: [
		      	CKEDITOR.dialog.okButton( editor, {
					label: 'OkCK',
					style : 'display:none;'
				}),
		      	CKEDITOR.dialog.cancelButton,

		      	{
                            id: 'myokbtnid',
                            type: 'button',
                            label: 'OK',
                            title: lang.validModif,
                            accessKey: 'C',
                            disabled: false,
                            onClick: function()
                                    {
                                        // code on click
                                        ClickOkBtn(this.getDialog());
                                    }
		      	}
		      ],
		// Dialog window contents definition.
		contents: [
			{
				// Definition of the Basic Settings dialog (page).
				id: 'slideshowinfoid',
				label: 'Basic Settings',
				align : 'center',
				// The tab contents.
				elements: [
                                        {
                                            type : 'text',
                                            id : 'id',
                                            style : 'display:none;',
                                            onLoad : function()
                                            {
                                                this.getInputElement().setAttribute( 'readOnly', true );
                                            }
                                        },
                                        {
                                            type: 'text',
                                            id: 'txturlid',
                                            style : 'display:none;',
                                            label: lang.imgList,
                                            onChange: function() {
                                                var dialog = this.getDialog(),
                                                    newUrl = this.getValue();
                                                if ( newUrl.length > 0 ) { //Prevent from load before onShow
                                                    var preview = dialog.previewImage;
                                                    preview.on( 'load', onImgLoadEvent, dialog );
                                                    preview.on( 'error', onImgLoadErrorEvent, dialog );
                                                    preview.on( 'abort', onImgLoadErrorEvent, dialog );
                                                    preview.setAttribute( 'src', newUrl );
                                                }
                                            }
                                        },
					{
                                            type : 'button',
                                            id : 'browse',
                                            hidden : 'true',
                                            style : 'display:inline-block;margin-top:0px;',
                                            filebrowser :
                                            {
                                                action : 'Browse',
                                                target: 'slideshowinfoid:txturlid',
                                                url: editor.config.filebrowserImageBrowseUrl || editor.config.filebrowserBrowseUrl
                                            },
                                            label : lang.imgAdd
					},

//					{
//						type : 'button',
//						id : 'browseDir',
//						style : 'display:inline-block;margin-top:0px;',
//						label : "toto",
//						onClick :  function() {
//							previewSlideShow(this.getDialog());
//						}
//					},

					{
					type: 'vbox',
                                        align: 'center',
					children: [
								{
									type: 'html',
									align : 'center',
									id: 'framepreviewtitleid',
									style: 'font-family: Amaranth; color: #1E66EB;	font-size: 20px; font-weight: bold;',
									html: lang.previewMode
								},
								{
									type: 'html',
									id: 'framepreviewid',
									align : 'center',
									style : 'width:500px;height:320px',
									html: ''
								},
								{
									type: 'hbox',
									id: 'imgparamsid',
									style : 'display:none;width:500px;',
									height: '325px',
									children :
										[
											{
												type : 'vbox',
												align : 'center',
												width : '400px',
												children :
												[
													{
														type : 'text',
														id : 'imgtitleid',
														label : lang.imgTitle,
														onChange: function() {
                                                                                                                    updateTitle(this.getDialog(), this.getValue());
														},
														onBlur: function() {
                                                                                                                    updateTitle(this.getDialog(), this.getValue());
														}
													},
													{
														type : 'text',
														id : 'imgdescid',
														label : lang.imgDesc,
														onChange: function() {
                                                                                                                    updateDescription(this.getDialog(), this.getValue());
														},
														onBlur: function() {
                                                                                                                    updateDescription(this.getDialog(), this.getValue());
														}
													},
													{
														type : 'html',
														id : 'imgpreviewid',
														style : 'width:400px;height:200px;',
														html: '<div>xx</div>'
													}
												]
											}
										]
								},
								{
								type : 'hbox',
                                                                align: 'center',
                                                                height: 110,
								widths: [ '25%', '50%'],
								children :
								[
				                    {
										type : 'vbox',
										children :
										[
											{
												type : 'checkbox',
												id : 'autostartid',
												label : lang.autoStart,
												'default' : 'checked',
												style : 'margin-top:15px;',
												onChange : commitValue,
												commit : commitValue,
												setup : loadValue
											},
											{
												type : 'checkbox',
												id : 'showtitleid',
												label : lang.showTitle,
												'default' : 'checked',
												onChange : commitValue,
												commit : commitValue,
												setup : loadValue
											},
											{
												type : 'checkbox',
												id : 'showcontrolid',
												label : lang.showControls,
												'default' : 'checked',
												onChange : commitValue,
												commit : commitValue,
												setup : loadValue
											},
											{
												type : 'checkbox',
												id : 'showthumbid',
												label : lang.showThumbs,
												'default' : 'checked',
					                    		onChange : commitValue,
												commit : commitValue,
												setup : loadValue
											},
											{
												type : 'checkbox',
												id : 'openOnClickId',
												label : lang.openOnClick,
												'default' : 'checked',
												onChange : commitValue,
												commit : commitValue,
												setup : loadValue
											}
						                ]
				                    },
								{
			                        type: 'select',
			                        id: 'imglistitemsid',
			                        label: lang.picturesList,
			                        multiple: false,
                                                style : 'height:125px;width:250px',
			                        items: [],
			                    	onChange : function( api ) {
			                    		//unselectIfNotUnique(this);
			                    		selectFirstIfNotUnique(this);
			                    	}
			                    },
			                    {
								type : 'vbox',
								children :
								[
									{
										type : 'button',
										id : 'previewbtn',
										style : 'margin-top:15px;margin-left:25px;',
										label : lang.previewMode,
										onClick :  function() {
											previewSlideShow(this.getDialog());
										}
									},
									{
										type : 'button',
										id : 'removeselectedbtn',
										style : 'margin-left:25px;',
										//style : 'display:none;',
										label : lang.imgDelete,
										onClick :  function() {
											removeSelected(this.getDialog());
										}
									},
									{
										type : 'button',
										id : 'editselectedbtn',
										style : 'margin-left:25px;',
										//style : 'display:none;',
										label : lang.imgEdit,
										onClick :  function() {
											editeSelected(this.getDialog());
										}
									},
									{
										type : 'hbox',
										children :
										[
											{
												type : 'button',
												id : 'upselectedbtn',
												style : 'width:32px; margin-left:25px;',
												//style : 'display:none;',
												label : lang.arrowUp,
												onClick :  function() {
													upDownSelected(this.getDialog(), -1);
												}
											},
											{
												type : 'button',
												id : 'downselectedbtn',
												style : 'width:32px;',
												//style : 'margin-left:5px;',
												//style : 'display:none;',
												label : lang.arrowDown,
												onClick :  function() {
													upDownSelected(this.getDialog(), 1);
												}
											}
										]
									}
								 ]
			                    }
			                ]
						},
	                    {
							type : 'hbox',
							children :
							[
//								{
//									type : 'text',
//									id : 'pictWidthtid',
//									label : lang.pictWidth,
//									maxLength : 3,
//									style : 'width:100px;',
//									'default' : '300',
//			                    	onChange : function( api ) {
//										var intRegex = /^\d+$/;
//										if(intRegex.test(this.getValue()) == false) {
//											console.log("setValue0: ", this.getValue());
//			                    			this.setValue(300);
//			                    		} else {
//											console.log("setValue1: ", this.getValue());
//			                    		}
//			                    		this.getDialog().params.updateVal(this.id, this.getValue());
//			                    		displaySelected(this.getDialog());
//			                    	},
//									commit : commitValue,
//									setup : loadValue,
//								},
								{
									type : 'text',
									id : 'pictheightid',
									label : lang.pictHeight,
									maxLength : 3,
									style : 'width:100px;',
									'default' : '300',
			                    	onChange : function( api ) {
										var intRegex = /^\d+$/;
										if(intRegex.test(this.getValue()) == false) {
			                    			this.setValue(300);
			                    		}
			                    		this.getDialog().params.updateVal(this.id, this.getValue());
			                    		displaySelected(this.getDialog());
			                    	},
									commit : commitValue,
									setup : loadValue
								},
								{
									type : 'text',
									id : 'speedid',
									label : lang.displayTime,
									maxLength : 3,
									style : 'width:100px;',
									'default' : '5',
			                    	onChange : function( api ) {
                                                        var intRegex = /^\d+$/;
                                                        if(intRegex.test(this.getValue()) == false) {
			                    			this.setValue(5);
			                    		}
			                    		this.getDialog().params.updateVal(this.id, this.getValue());
			                    		displaySelected(this.getDialog());
			                    	},
									commit : commitValue,
									setup : loadValue
								},
								{
									type : 'text',
									id : 'animspeedid',
									label : lang.transitionTime,
									style : 'width:100px;',
									maxLength : 4,
									'default' : '500',
			                    	onChange : function( api ) {
                                                        var intRegex = /^\d+$/;
                                                        if(intRegex.test(this.getValue()) == false) {
			                    			this.setValue(500);
			                    		}
			                    		this.getDialog().params.updateVal(this.id, this.getValue());
			                    		displaySelected(this.getDialog());
			                    	},
									commit : commitValue,
									setup : loadValue
								},
								{
									type : 'select',
									id : 'transitiontypeid',
									label : lang.transition,
									  // add-gallery effects 'slide-vert', 'resize', 'fade', 'none' or false
									  // effect: 'slide-hori',
									items : [ [ lang.tr1, 'none' ], [ lang.tr2, 'resize' ], [ lang.tr3, 'slide-vert' ], [ lang.tr4, 'slide-hori' ], [lang.tr5, 'fade'] ],
									'default' : 'resize',
									style : 'width:100px;',
									commit : commitValue,
									setup : loadValue,
									onChange : commitValue
								}
							]
	                    }
			            ]
					}
				]
			}
		],


		onLoad: function() {
		},
		// Invoked when the dialog is loaded.
		onShow: function() {
			this.dialog = this;
			this.slideshowDOM = null;
			this.openCloseStep = true;
			this.fakeImage =  null;
			var slideshowDOM = null;
			this.imagesList = [];
			this.params = [];
			// To get dimensions of poster image
			this.previewImage = editor.document.createElement( 'img' );
			this.okRefresh = true;

                        var fakeImage = this.getSelectedElement();
			if ( fakeImage && fakeImage.data( 'cke-real-element-type' ) && fakeImage.data( 'cke-real-element-type' ) == 'slideShow' )
			{
				this.fakeImage = fakeImage;
				slideshowDOM = editor.restoreRealElement( fakeImage );
			}

			// Create a new <slideshow> slideshowDOM if it does not exist.
			if ( !slideshowDOM) {
				this.params.push(['slideshowid', numbering( 'slideShow' )]);

				// Insert placeHolder image
				initImgListFromFresh(this);
				// Invoke the commit methods of all dialog elements, so the dialog.params array get Updated.
				this.commitContent(this);
//				console.log( "Params New -> " + this.params );
//				console.log( "Images New -> " + this.imagesList );
			} else {
				this.slideshowDOM = slideshowDOM;
				// Get the  reference of the slideSjow Images Container
				var slideShowContainer =  getImagesContainerBlock(this, slideshowDOM);
				if (slideShowContainer == null) {
					alert("BIG Problem slideShowContainer !!");
					return false;
				}
				var slideshowid = slideshowDOM.getAttribute('id');
				if (slideshowid == null) {
					alert("BIG Problem slideshowid !!");
					return false;
				}
				this.params.push(['slideshowid', slideshowid]);
				// a DOM has been found updatet images List and Dialog box from this DOM
				initImgListFromDOM(this, slideShowContainer);
				// Init params Array from DOM
				// Copy all attributes to an array.
				var domDatas = slideshowDOM.$.dataset;
                                var param;
				for ( param in  domDatas ) {
                                    this.params.push( [ param, domDatas[ param ] ] );
                                }

				// Invoke the setup methods of all dialog elements, to set dialog elements values with DOM input data.
				this.setupContent(this, true);
				//updateFramePreview(this);
				this.newSlideShowMode = false;
//				console.log( "Params Old -> " + this.params );
//				console.log( "Images Old -> " + this.imagesList );
			}
			this.openCloseStep = false;
			previewSlideShow(this);
		},

		// This method is invoked once a user clicks the OK button, confirming the dialog.
		// I just will return false, as the real OK Button has been redefined
		//  -This was the only way I found to avoid dialog popup to close when hitting the keyboard "ENTER" Key !!
		onOk: function() {
//			var okr = this.okRefresh;
//			if (this.okRefresh == true) {
//				console.log('OKOKOK 0 :'+this.okRefresh);
//				this.okRefresh = false;
//				this.commitContent(this);
//				myVar = setTimeout(
//						function(obj){
//									obj.okRefresh = true;
//									},500, this);
//			}
			return false;
		},

		onHide: function() {
			cleanAll(this);
		}
	};
});
