/**
 * Copyright (c) 2014, CKSource - Frederico Knabben. All rights reserved.
 * Licensed under the terms of the MIT License (see LICENSE.md).
 *
 * Plugin to add CSS animations on elements into the CKEditor editing area.
 *
 */
/*
JSLint Global variables :
CKEDITOR
console
editor
document
navigator
alert
CSSRule
XMLHttpRequest
parseCSS3AnimationShorthand
window
lang
setTimeout
 */
/* Default Tags (the one I selected from html5 standard)
 * Inline html tags elements
 * <img> <span> <button> <label>
 * Block html tags elements
 * <div> <figure> <form> <h1> <h2> <h3> <h4> <h5> <h6> <p> <section> <video>
 * <tr> <td>
 */
// css = cssSelectorStart
// cds = cssDefStart
// cso = cssSelectorOver
// cdo = cssDefOver
// csc = cssSelectorClick
// cdc = cssDefClick
// ral = css Remove after Load flag
(function () {
    "use strict";

    var allowedAnimsDef = {};
    allowedAnimsDef.BOUNCE = ['bounce', 'bounceIn', 'bounceInDown', 'bounceInLeft', 'bounceInRight', 'bounceInUp', 'bounceOut', 'bounceOutDown', 'bounceOutLeft', 'bounceOutRight', 'bounceOutUp'];
    allowedAnimsDef.FADE = ['fadeIn', 'fadeInDown', 'fadeInDownBig', 'fadeInLeft', 'fadeInLeftBig', 'fadeInRight', 'fadeInRightBig', 'fadeInUp', 'fadeInUpBig', 'fadeOut', 'fadeOutDown', 'fadeOutDownBig', 'fadeOutLeft', 'fadeOutLeftBig', 'fadeOutRight', 'fadeOutRightBig', 'fadeOutUp', 'fadeOutUpBig'];
    allowedAnimsDef.FLIP = ['flip', 'flipInX', 'flipInY', 'flipOutX', 'flipOutY'];
    allowedAnimsDef.ROTATE = ['rotate', 'rotateIn', 'rotateInDownLeft', 'rotateInDownRight', 'rotateInUpLeft', 'rotateInUpRight', 'rotateOut', 'rotateOutDownLeft', 'rotateOutDownRight', 'rotateOutUpLeft', 'rotateOutUpRight', 'rotateRound'];
    allowedAnimsDef.SLIDE = ['slideInDown', 'slideInLeft', 'slideInRight', 'slideInUp', 'slideOutDown', 'slideOutLeft', 'slideOutRight', 'slideOutUp'];
    allowedAnimsDef.ZOOM = ['zoomIn', 'zoomInDown', 'zoomInLeft', 'zoomInRight', 'zoomInUp', 'zoomOut', 'zoomOutDown', 'zoomOutLeft', 'zoomOutRight', 'zoomOutUp'];
    allowedAnimsDef.OTHER = ['flash', 'hinge', 'jello', 'lightSpeedIn', 'lightSpeedOut', 'pulse', 'rollIn', 'rollOut', 'rubberBand', 'shake', 'swing', 'tada', 'wobble'];

    var allowedTagsDef = ['img', 'span', 'button', 'label', 'div', 'figure', 'form', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'p', 'section', 'video', 'table', 'tr', 'td'];

//    function getBrowser() {
//		var browserName;
//		var nAgt = navigator.userAgent;
//		var verOffset;
////		console.log(navigator);
//		if ((verOffset=nAgt.indexOf("OPR/"))!=-1) {
//			 browserName = "Opera";
//			} else if ((verOffset=nAgt.indexOf("Opera"))!=-1) {
//			 browserName = "Opera";
//			} else if ((verOffset=nAgt.indexOf("Maxthon"))!=-1) {
//			 browserName = "Maxthon";
//			} else if ((verOffset=nAgt.indexOf("MSIE"))!=-1) {
//			 browserName = "Microsoft Internet Explorer";
//			} else if ((verOffset=nAgt.indexOf("Chrome"))!=-1) {
//			 browserName = "Chrome";
//			} else if ((verOffset=nAgt.indexOf("Safari"))!=-1) {
//			 browserName = "Safari";
//			} else if ((verOffset=nAgt.indexOf("Firefox"))!=-1) {
//			 browserName = "Firefox";
//			} else if ( (nameOffset=nAgt.lastIndexOf(' ')+1) < (verOffset=nAgt.lastIndexOf('/')) ) {
//			 browserName = nAgt.substring(nameOffset,verOffset);
//			 if (browserName.toLowerCase()==browserName.toUpperCase()) {
//			  browserName = navigator.appName;
//			 }
//			}
//		return browserName;
//    }

    function removeDomainFromUrl(string) {
        var str = string.replace(/^https?:\/\/[^\/]+/i, '');
        return str;
    }
//    var highLightedElem = null;
    //var highLightedElemBG = "";
    var animClassPrefix = "ckAnimClass_";
    var animScriptTitle = "_ckAnimScript_";
//    var idCnt = -1;
//    var cssAnimationObject = {};
    var customCssScriptTitle = "_ckAnimCustomCss_";
    var customCssFile = "";

    function myGetNamedItem(elm, name) {
    	var i;
    	if (elm.namedItem) {
    		return elm.namedItem(name);
    	} else {
        	for (i=0;i<elm.length;i++) {
        		if (elm.item(i).name == name) {
        			return elm.item(i);
        	 	}
        	 }
     		return null;
    	}
        alert ("Your browser doesn't support item or namedItem method.");
    }
    function generateUUID() {
        var d = new Date().getTime();
//        var uuid = 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        var uuid = 'xxxyyxxxxyyx'.replace(/[xy]/g, function(c) {
            var r = (d + Math.random()*16)%16 | 0;
            d = Math.floor(d/16);
            return (c=='x' ? r : (r&0x7|0x8)).toString(16);
        });
        return uuid;
    }
    function cmdHighLightTagElem(elem) {
        return {
            // It applies to a "block-like" context.
            // requiredContent: elem,
            exec: function () {
                CKEDITOR.plugins.cssanim.runHighLightElem(elem);
            }
        };
    }

    function cmdAddAnimTagElem(elem) {
        return {
            exec: function (editor) {
                CKEDITOR.plugins.cssanim.runAddAnimDialog(editor, elem);
            }
        };
    }

    function cmdEditAnimTagElem(elem) {
        return {
            exec: function (editor) {
                CKEDITOR.plugins.cssanim.runEditAnimDialog(editor, elem);
            }
        };
    }

    function cmdRemoveAnimTagElem(elem) {
        return {
            exec: function () {
                CKEDITOR.plugins.cssanim.runRemoveAnimElem(elem);
            }
        };
    }

    // Add left menu items for all html tags / leements from selected to top
    function addMenusItemsPath(rootPath, editor, element, path) {
        var elementsFound = false;
        var i;
        var elem;
        var tag;
        var subMenus;
        var lang = editor.lang.cssanim;

//        console.log("addMenusItemsPath path:", path.length, path);
        if (!element || element.isReadOnly()) {
            return elementsFound;
        }
        for (i = 0; i < path.length; i += 1) {
        	elem = path[i];
            tag = elem.getName().toLowerCase();
            if (editor.config.allowedTags.indexOf(tag) >= 0) {
                elementsFound = true;
                tag += i.toString();
                editor.addCommand('highlightTag' + tag, new CKEDITOR.command(editor, cmdHighLightTagElem(elem)));
                editor.addCommand('addAnimTag' + tag, new CKEDITOR.command(editor, cmdAddAnimTagElem(elem)));
                editor.addCommand('editAnimTag' + tag, new CKEDITOR.command(editor, cmdEditAnimTagElem(elem)));
                editor.addCommand('removeAnimTag' + tag, new CKEDITOR.command(editor, cmdRemoveAnimTagElem(elem)));
                //console.log("add command for :", tag);
                subMenus = {};
                subMenus['highlightsubmenu' + tag] = {
                    label: lang.HighLight,
                    group: "highlightGroup",
                    command: 'highlightTag' + tag,
                    icon: rootPath + 'images/lamp.png',
                    order: 2
                };
                subMenus['addanimsubmenu' + tag] = {
                    label: lang.AddAnimations,
                    group: "highlightGroup",
                    command: 'addAnimTag' + tag,
                    icon: rootPath + 'images/add.png',
                    order: 3
                };
                subMenus['editanimsubmenu' + tag] = {
                    label: lang.EditAnimations,
                    group: "highlightGroup",
                    command: 'editAnimTag' + tag,
                    icon: rootPath + 'images/edit.png',
                    order: 4
                };
                subMenus['removeanimsubmenu' + tag] = {
                    label: lang.RemoveAnimations,
                    group: "highlightGroup",
                    command: 'removeAnimTag' + tag,
                    icon: rootPath + 'images/del.png',
                    order: 4
                };
                //console.log("addMenusItemsPath:", subMenus);
                if (editor.addMenuItems) {
                    editor.addMenuItems(subMenus);
                }
            }
        }
        return elementsFound;
    }

    function getItemsFunc() {
        var obj = {};
        obj['highlightsubmenu' + this.tag] = CKEDITOR.TRISTATE_OFF;
        //if (this.elem.data('animation') == null) {
        if (this.elem.$.classList.toString().indexOf(animClassPrefix) < 0) {
            obj['addanimsubmenu' + this.tag] = CKEDITOR.TRISTATE_OFF;
            obj['editanimsubmenu' + this.tag] = CKEDITOR.TRISTATE_DISABLED;
            obj['removeanimsubmenu' + this.tag] = CKEDITOR.TRISTATE_DISABLED;
        } else {
            obj['addanimsubmenu' + this.tag] = CKEDITOR.TRISTATE_DISABLED;
            obj['editanimsubmenu' + this.tag] = CKEDITOR.TRISTATE_OFF;
            obj['removeanimsubmenu' + this.tag] = CKEDITOR.TRISTATE_OFF;
        }
        //console.log("OBJ", obj);
        return obj;
    }
    // Add listeners for menu items
    function listeners(editor, element, path) {
        //console.log("listeners:", path.length, path);
        var nodes = {};
        var elem, tag, label, i;
        if (!element || element.isReadOnly()) {
            return null;
        }
        var menusItems = {};
        for (i = 0; i < path.length; i += 1) {
            elem = path[i];
            tag = elem.getName().toLowerCase();
            label = tag.toUpperCase();
//            if (elem.dataset.ckeRealElementType === undefined) {
                if (editor.config.allowedTags.indexOf(tag) >= 0) {
                    tag += i.toString();
                    nodes['topMenu' + tag] = CKEDITOR.TRISTATE_OFF;
                    menusItems['topMenu' + tag] = {
                        label: label,
                        icon: CKEDITOR.basePath + 'plugins/cssanim/icons/cssanim.png',
                        group: 'highlightGroup',
                        tag: tag,
                        elem: elem,
                        order: 0,
                        getItems: getItemsFunc
                    };
                }
//            }
        }
        editor.addMenuItems(menusItems);
        return nodes;
    }
    // Register the plugin within the editor.
    CKEDITOR.plugins.add('cssanim', {
        // Register the icons.
        lang: 'en,fr',
        icons: 'cssanim',
    	requires: 'contextmenu',
        // onLoad: function(editor) {
        // //console.log("ON LOAD ------");
        // },
        beforeInit: function (/*editor*/) {
        	// add needed css files to editor
            //console.log("BEFORE INIT --- ", editor, CKEDITOR.document); // 'bar'
//            var brow = getBrowser();
        	CKEDITOR.document.appendStyleSheet(CKEDITOR.basePath + 'plugins/cssanim/css/cssanim.css');
//            CKEDITOR.document.appendStyleSheet(CKEDITOR.basePath + 'plugins/cssanim/css/tabcontent.css');

//            if ((brow !== 'Maxthon') && (brow !== 'Safari')) {
//                CKEDITOR.document.appendStyleSheet(CKEDITOR.basePath + 'plugins/cssanim/css/tabcontentnotest.css');
//    		} else {
//                CKEDITOR.document.appendStyleSheet(CKEDITOR.basePath + 'plugins/cssanim/css/tabcontentnotest.css');
//    		}
        },
        afterInit: function (editor) {
            //console.log("AFTER INIT --- ", editor);
            var dataProcessor = editor.dataProcessor;
            var htmlFilter = dataProcessor && dataProcessor.htmlFilter;
//                dataFilter = dataProcessor && dataProcessor.dataFilter;
//            if (dataFilter) {
//                dataFilter.addRules({
//                });
//            }
            if (htmlFilter) {
                //console.log("htmlFilter:", htmlFilter);
                htmlFilter.addRules({
                    elements: {
                        $: function (element) {
                        	//console.log("htmlFilter", element);
                        	// Called for each element before generating the html output.
                            // If one element is highlighted, then remove the highlight style to avoid to push it in the created html
                            //if (element.attributes.class) {
                                if (element.hasClass('highlight_tag')) {
                                    //console.log("htmlFilter -> element:", element);
                                    //element.attributes.class = element.attributes.class.replace('highlight_tag', '');
                                	element.removeClass('highlight_tag');
                                	var style = new CKEDITOR.htmlParser.cssStyle(element.attributes.style);
                                    style.rules["background-color"] = element.attributes['data-animation-bg'];
                                    style.rules["outline"] = element.attributes['data-animation-outline'];
                                    element.attributes.style = style.toString();
                                    // reset highlight
                                    delete element.attributes["data-animation-bg"];
                                    delete element.attributes["data-animation-outline"];
                                }
                            //}
                        }
                    }
                });
            }
        },
        // The plugin initialization logic goes inside this method.
        init: function (editor) {
            //console.log("INIT ------", editor);
            // Get Config & Lang
            var lang = editor.lang.cssanim;
            var rootPath = this.path,
                defaultConfig = {
            		// default accepted tags, can be overriden by config file
                    acceptedTags: allowedTagsDef,
            		// default available animations, can be overriden by config file
                    acceptedAnimations: allowedAnimsDef,
                    highlightBGColor: '#87CEFA',
                    highlightBorder: '3px',
                    highlightPadding: '3px'
                };
            var config = CKEDITOR.tools.extend(defaultConfig, editor.config.cssanim || {}, true);
            editor.config.onLoadAllowedTags = config.acceptedTags.sort();
            editor.config.onLoadAllowedAnimations = config.acceptedAnimations;

            editor.config.allowedTags = editor.config.onLoadAllowedTags;	// may be filtered in main dialog
            editor.config.allowedAnimations = editor.config.onLoadAllowedAnimations;
            editor.config.highlightBGColor = config.highlightBGColor;	// may be changed in main dialog
            editor.config.highlightBorder = config.highlightBorder;		// may be changed in main dialog
            editor.config.highlightPadding = config.highlightPadding;	// may be changed in main dialog
            editor.config.customCssFilePath = config.customCssFilePath;	// may be changed in main dialog
            //
            // Functions
            //
            // Define an editor command that opens our dialog window.
            var cssanimAddAnim = new CKEDITOR.dialogCommand('cssanimAddAnimDialog');
            editor.addCommand('_myLaunchAddAnimDialog', cssanimAddAnim);
            var cssanimMain = new CKEDITOR.dialogCommand('cssanimMainDialog', {
                // Allow ckAnimClass class with data-aniomation attributes.
            	// Addinionaly allowed 'button' and 'span' (as they seems to be filtered by ckeditor default)
                allowedContent: 'button;span;div{*};*[id,data-animation*](ckAnimClass*)'
            });
            editor.addCommand('cssanim', cssanimMain);
            // Create a toolbar button that executes the above command.
            editor.ui.addButton('cssanim', {
                // The text part of the button (if available) and the tooltip.
                label: lang.genProperties,
                // The command to execute on click.
                command: 'cssanim',
                // The button placement in the toolbar (toolbar group name).
                toolbar: 'insert'
            });
            editor.addMenuGroup('highlightGroup', 200);
            if (editor.contextMenu) {
                // Add a context menu group with the Edit cssanimeviation item.
                editor.addMenuGroup('cssanimGroup');
                editor.contextMenu.addListener(function (element /*, selection*/) {
                    //console.log("SELECTION", element, selection);
                    var path = CKEDITOR.plugins.cssanim.getPathToTop(element);
                    //console.log("PATH TO TOP", path);
                    var foundSome = addMenusItemsPath(rootPath, editor, element, path);
                    if (foundSome === true) {
                        return listeners(editor, element, path);
                    } else {
                        return null;
                    }
                });
            }
            // Register our dialog file -- this.path is the plugin folder path.
            CKEDITOR.dialog.add('cssanimMainDialog', this.path + 'dialogs/cssanim.min.js');
            CKEDITOR.dialog.add('cssanimAddAnimDialog', this.path + 'dialogs/cssaddanim.min.js');
            editor.on('contentDom', function (/*e*/) {
                // Ini document editor
                CKEDITOR.plugins.cssanim.init(editor);
                if (editor.config.customCssFilePath) {
	                 CKEDITOR.plugins.cssanim.getCustomCss(editor.config.customCssFilePath);
                }
            });
            editor.on('toHtml', function (evt) {
            	//console.log("ON toHtml 10");
                // Called when loading the html inside the text area
                var key;
                var source, re, m;
                // Here we remove the javascript function removeAnimation(event)
                // The function will be re-created if needed
                for (key in evt.data.dataValue.children) {
                    if (evt.data.dataValue.children[key].type === CKEDITOR.NODE_COMMENT) {
                        if ((evt.data.dataValue.children[key].value.indexOf(animScriptTitle) >= 0) || (evt.data.dataValue.children[key].value.indexOf(customCssScriptTitle) >= 0)) {
                            if (evt.data.dataValue.children[key].value.indexOf(customCssScriptTitle) >= 0) {
                                // A custom CSS file was defined, So init the correct values fot it !!
                                source = decodeURIComponent(evt.data.dataValue.children[key].value.replace(/^\{cke_protected\}/, ''));
                                re = /^.*(customCss\.setAttribute\("href",)(.*)\);customCss(.*)/;
                                //if ((m = re.exec(source)) !== null) {
                                m = re.exec(source);
                                if (m !== null) {
                                    if (m.index === re.lastIndex) {
                                        re.lastIndex += 1;
                                    }
                                    //console.log("==================", m[2]);
                                    editor.config.customCssFilePath = m[2];
                                    editor.config.customCssFilePath = editor.config.customCssFilePath.replace(/"/gi, '');
                                    editor.config.customCssFilePath = editor.config.customCssFilePath.trim();
                                    //console.log("==================", editor.config.customCssFilePath);
                                }
                            }
                            evt.data.dataValue.children[key].remove();
                        }
                    }
                }
            }, null, null, 10);
            editor.on('toDataFormat', function (evt) {
                CKEDITOR.plugins.cssanim.cleanHighlight();
                var pluginPath = removeDomainFromUrl(CKEDITOR.plugins.get('cssanim').path);
                var cssStr0 = "\"";
                var rmFuncStr = "";
                var foundOneAnim = false;
                var obj0;
                var removeAfterLoadFunc = false;
            	var cssAnimationObject = CKEDITOR.plugins.cssanim.getAllAnimElem();

                for (obj0 in cssAnimationObject) {
                    if (cssAnimationObject[obj0].css) {
                        foundOneAnim = true;
                        cssStr0 += cssAnimationObject[obj0].css;
                        cssStr0 += cssAnimationObject[obj0].cds;
                        if (cssAnimationObject[obj0].ral) {
                            // remove after Load flag is ON, create the javascript listener for end of animation
                        	// and the javascript function to remove
                            removeAfterLoadFunc = true;
                            rmFuncStr += "document.getElementsByClassName('" + animClassPrefix + obj0 + "Start')[0].addEventListener('animationend', removeAnimation); ";
                            rmFuncStr += "document.getElementsByClassName('" + animClassPrefix + obj0 + "Start')[0].addEventListener('webkitAnimationEnd', removeAnimation); ";
                        }
                        // Debug staff
                        // rmFuncStr += "console.log('Added Anim for: ',document.getElementsByClassName('" + animClassPrefix + obj0 + "Start')[0]); ";
                    }
                    if (cssAnimationObject[obj0].cso) {
                        foundOneAnim = true;
                        cssStr0 += cssAnimationObject[obj0].cso;
                        cssStr0 += cssAnimationObject[obj0].cdo;
                    }
                    if (cssAnimationObject[obj0].csc) {
                        foundOneAnim = true;
                        cssStr0 += cssAnimationObject[obj0].csc;
                        cssStr0 += cssAnimationObject[obj0].cdc;
                    }
                }
                cssStr0 += "\"";
                if (foundOneAnim || (customCssFile !== "")) {
                	// foud at least on animation or a custom defined css file, so put in in the generatde document
                    var removeAnimFunc = "";
                    var s1 = window.location.href;
                    var s2 = s1.replace(s1.split("/").pop(),"");
                    var isAbsolute = new RegExp('^([a-z]+://|//)');
                    if (removeAfterLoadFunc) {
                        // remove after Load flag is ON, createthe javascript function to remove the animation after completion
                        removeAnimFunc += "function removeAnimation(event) {";
//                        removeAnimFunc += "console.log(\"Animation End: \", event);";
                        removeAnimFunc += "var thisId = event.currentTarget.getAttribute('id');";
//                        removeAnimFunc += "console.log(\"Animation End: \", thisId);";
                        removeAnimFunc += "this.classList.remove('" + animClassPrefix + "'+thisId+'Start');";
                        removeAnimFunc += "this.removeEventListener('animationend', removeAnimation);";
                        removeAnimFunc += "}";
                    }
                    var injectCssStr = 'var cssStyle = document.createElement("style");';
                    var injectCssAnimStr = '';
                    if (isAbsolute.test(customCssFile) === false) {
                    	customCssFile = s2 + customCssFile;
                    }
	                if (foundOneAnim) {
	                    // add prefixfree
	                    var scriptPrefixFree = new CKEDITOR.htmlParser.element('script', {
	                        src: pluginPath + 'css/prefixfree.min.js'
	                    });
	                    evt.data.dataValue.add(scriptPrefixFree);
                        injectCssStr += 'cssStyle.innerHTML = ' + cssStr0 + ';';
                        injectCssStr += 'document.getElementsByTagName("head")[0].appendChild(cssStyle);';
                        injectCssAnimStr += 'var fileref = document.createElement("link");';
                        injectCssAnimStr += 'fileref.setAttribute("rel", "stylesheet");';
                        injectCssAnimStr += 'fileref.setAttribute("type", "text/css");';
                        injectCssAnimStr += 'fileref.setAttribute("href", "' + pluginPath + 'css/cssanim.css");';
                        injectCssAnimStr += 'document.getElementsByTagName("head")[0].appendChild(fileref);';
                    }
                    if (customCssFile !== "") {
                        injectCssAnimStr += 'var customCss = document.createElement("link");';
                        injectCssAnimStr += 'customCss.setAttribute("rel", "stylesheet");';
                        injectCssAnimStr += 'customCss.setAttribute("type", "text/css");';
                        injectCssAnimStr += 'customCss.setAttribute("href", "' + customCssFile + '");';
                        injectCssAnimStr += 'customCss.setAttribute("title", "' + customCssScriptTitle + '");';
                        injectCssAnimStr += 'document.getElementsByTagName("head")[0].appendChild(customCss);';
                    }
                    var scriptAddCssAnimations = new CKEDITOR.htmlParser.element('script', {
                        title: animScriptTitle
                    });
                    var cssAnimationsFunction = '\n' + "(function(){ " + removeAnimFunc + injectCssStr + injectCssAnimStr + rmFuncStr + "})();";
                    scriptAddCssAnimations.setHtml(cssAnimationsFunction);
                    evt.data.dataValue.add(scriptAddCssAnimations);
                }
            }, null, null, 14);
        }
    });
    CKEDITOR.on('instanceReady', function (e) {
        //console.log("---------------- CKEDITOR instanceReady ---------------");
        var head = document.getElementsByTagName("head")[0];
        var js = document.createElement('script');
        js.src = CKEDITOR.basePath + 'plugins/cssanim/parseCSS3AnimationShorthand.min.js';
        js.type = 'text/javascript';
        head.appendChild(js);
        e.editor.document.$.addEventListener("contextmenu", function (/*event*/) {
            //console.log("----------------------contextmenu--------------------------", event);
        	CKEDITOR.plugins.cssanim.cleanHighlight();
        });
    });
//    CKEDITOR.on('contentDom', function ( /*event*/ ) {
//        //console.log("---------------- CKEDITOR contentDom ---------------");
//    });
//    CKEDITOR.on('loaded', function ( /*event*/ ) {
//        //console.log("---------------- CKEDITOR loaded ---------------");
//    });
//    CKEDITOR.on('pluginsLoaded', function ( /*event*/ ) {
//        //console.log("---------------- CKEDITOR pluginsLoaded ---------------");
//    });
    CKEDITOR.plugins.cssanim = {
        ckEditor: null,
        documentEditor: null,
        //    cssanimAddAnimDialog : null,
        curSelectedElement: null,
        managePending: false,
        pendingChanges: [],	// will store the change done in addAnim dialog in case of a cancel in main dialog (undo list)
        cssAnimDialog: null,
        allAvailableAnimsCssAnim: [],
        allAvailableAnimsCustom: [],
        allAvailableAnimsConflict: [],
        init: function (editor) {
            this.ckEditor = editor;
            this.documentEditor = editor.document;
            var i, j;
            // get cssanim.css style sheet
            for (i = document.styleSheets.length - 1; i >= 0; i -= 1) {
                if ((document.styleSheets[i].href) && (document.styleSheets[i].href.indexOf("cssanim.css") >= 0)) {
                    for (j = 0; j < document.styleSheets[i].cssRules.length; j += 1) {
                        if ((document.styleSheets[i].cssRules[j].type === CSSRule.KEYFRAMES_RULE)
                        	|| (document.styleSheets[i].cssRules[j].type === CSSRule.WEBKIT_KEYFRAMES_RULE)) {
                            this.allAvailableAnimsCssAnim.push(document.styleSheets[i].cssRules[j].name);
                        }
                    }
                }
            }
        },
        setHighLightBgColor: function (color) {
            this.ckEditor.config.highlightBGColor = color;
        },
        setHighLightBorder: function (b) {
            this.ckEditor.config.highlightBorder = b +'px';
        },
        cleanHighlight: function () {
        	var highLightedElem = this.documentEditor.$.getElementsByClassName('highlight_tag');
        	var i, elem;
            for (i=0; i<highLightedElem.length; i+=1) {
            	elem = highLightedElem[i];
                elem.classList.remove('highlight_tag');
                elem.style.backgroundColor =  elem.getAttribute("data-animation-bg");
                elem.style.outline = elem.getAttribute("data-animation-outline");
                elem.getAttribute("data-animation-bg");
                elem.getAttribute("data-animation-outline");
            }
        },
       getSurroundElem: function (editor, elem) {
            var elemAsc = editor.getSelection().getStartElement().getAscendant(elem, true);
            return elemAsc;
        },
        setCssAnimDialog: function (dialog) {
            //console.log("setCssAnim", dialog);
            this.cssAnimDialog = dialog;
        },
        refreshCssAnimDialogAnimationsTab: function () {
            this.cssAnimDialog.selectPage('tab-allowed-tags');
            this.cssAnimDialog.selectPage('tab-doc-animations');
        },
        runHighLightElem: function (elem) {
        	//console.log('runHighLightElem', elem);
        	CKEDITOR.plugins.cssanim.cleanHighlight();
            elem.setAttribute("data-animation-bg", elem.getStyle('backgroundColor'));
            elem.setAttribute("data-animation-outline", elem.getStyle('outline'));
            elem.setStyle('backgroundColor', this.ckEditor.config.highlightBGColor);
            elem.setStyle('outlineWidth', this.ckEditor.config.highlightBorder);
            elem.setStyle('outlineStyle', 'solid');
            elem.setStyle('outlineColor', 'red');
            elem.addClass('highlight_tag');
        },
        runHighLightElemById: function (id) {
        	CKEDITOR.plugins.cssanim.cleanHighlight();
            var elem = CKEDITOR.plugins.cssanim.documentEditor.getById(id);
           if (!elem) {
                return;
            }
           CKEDITOR.plugins.cssanim.runHighLightElem(elem);
        },
        getAllAnimElem: function() {
        	var elems = this.documentEditor.$.getElementsByClassName('ckAnimClass_');
        	var cssAnimationObject = {};
            var obj, element;
            var i;
            for (i=0; i <elems.length; i+=1) {
            	element = elems[i];
	            obj = element.getAttribute('data-animation');
	            obj = decodeURIComponent(obj);
	            obj = JSON.parse(obj);
	            obj.elem = element;
	            cssAnimationObject[element.id] = obj;
            }
            return cssAnimationObject;
        },
        runAddAnimDialogById: function (id) {
            var elem = CKEDITOR.plugins.cssanim.documentEditor.getById(id);
            if (!elem) {
                return;
            }
            this.curSelectedElement = elem;
            this.managePending = true;
            this.ckEditor.execCommand("_myLaunchAddAnimDialog");
        },
        // Show the "add animations" dialog
        runAddAnimDialog: function (editor, elem) {
        	CKEDITOR.plugins.cssanim.cleanHighlight();
            //        editor.curSelectedElement = elem;
            this.curSelectedElement = elem;
            this.managePending = false;
            editor.execCommand("_myLaunchAddAnimDialog");
        },
        // Show the "add animations" dialog
        runAddAnimElem: function (elem, res) {
            var animStart = res.animStart;
            var animClick = res.animClick;
            var animOver = res.animOver;
            var ral = res.ral;
            var data;
            var cssObj;
            if (animStart || animClick || animOver) {
            	// FAKE ELEMENTS OLD CODE
                if (elem.getId() === null) {
                    elem.setAttribute('id', generateUUID());
                }
                // Remove anim class in case it already exists
                elem.removeClass(animClassPrefix);
                elem.removeClass(animClassPrefix + elem.getId());
                elem.removeClass(animClassPrefix + elem.getId() + 'Start');
                // add the new one
                elem.addClass(animClassPrefix);
                elem.addClass(animClassPrefix + elem.getId());
                cssObj = {};
                cssObj.id = elem.getId();
                if (animStart) {
                    elem.addClass(animClassPrefix + elem.getId() + 'Start');
                    cssObj.css = '.' + animClassPrefix + elem.getId() + 'Start';
                    cssObj.cds = "{animation: " + animStart + ";}";
                    cssObj.ral = ral;
                }
                if (animOver) {
                    cssObj.cso = '.' + animClassPrefix + elem.getId() + ':hover';
                    if ((animOver.indexOf('cssAnimPause') >= 0) || (animOver.indexOf('animation-play-state') >= 0)) {
                    	cssObj.cdo = "{animation-play-state: paused" + ";}";
                    } else {
                        cssObj.cdo = "{animation: " + animOver + ";}";
                    }
                }
                if (animClick) {
                    cssObj.csc = '.' + animClassPrefix + elem.getId() + ':active';
                    if ((animClick.indexOf('cssAnimPause') >= 0) || (animClick.indexOf('animation-play-state') >= 0)) {
                    	cssObj.cdc = "{animation-play-state: paused" + ";}";
                    } else {
                        cssObj.cdc = "{animation: " + animClick + ";}";
                    }
                }
                //console.log("runAddAnimElem cssObj ------------>", cssObj);
                data = encodeURIComponent(JSON.stringify(cssObj));
                elem.setAttribute("data-animation", data);
            } else {
                // no more animation, so remove them all
                //console.log("REMOVE ALL ANIMATIONS ->", elem);
                CKEDITOR.plugins.cssanim.runRemoveAnimElem(elem);
            }
        },
        // Restore animation, called if any pending changes need to be restored
        restoreAnimOnElemById: function (id, anim) {
            //console.log("RESTORE ANIMATIONS ->", id);
            var elem = CKEDITOR.plugins.cssanim.documentEditor.getById(id);
            var res = {};
            var exp;
            res.animStart = null;
            res.animClick = null;
            res.animOver = null;
            res.ral = anim.ral;
            if (anim.cds) {
                exp = anim.cds.replace('{animation:', '').replace('}', '').replace(';', '').trim();
                res.animStart = exp;
            }
            if (anim.cdo) {
                exp = anim.cdo.replace('{animation:', '').replace('}', '').replace(';', '').trim();
                res.animOver = exp;
            }
            if (anim.cdc) {
                exp = anim.cdc.replace('{animation:', '').replace('}', '').replace(';', '').trim();
                res.animClick = exp;
            }
            //console.log("restoreAnimOnElemById RES =", res);
            CKEDITOR.plugins.cssanim.runAddAnimElem(elem, res);
        },
        // Open the Add Animation Dialog
        runEditAnimDialog: function (editor, elem) {
            //console.log("runEditAnimDialog exec:", elem);
        	CKEDITOR.plugins.cssanim.cleanHighlight();
            this.curSelectedElement = elem;
            editor.execCommand("_myLaunchAddAnimDialog");
        },
        runRemoveAnimElem: function (elem) {
            //console.log("runRemoveAnimElem exec:", elem);
        	CKEDITOR.plugins.cssanim.cleanHighlight();
            //console.log("runRemoveAnimElem -> ", elem.classList);
        	  // FAKE ELEMENTS OLD CODE
//            if (elem.dataset.ckeRealElementType !== undefined) {
//                //        var realElem = CKEDITOR.dom.element.createFromHtml(decodeURIComponent(
//                //            elem.dataset.ckeRealelement ), this.document );
//                //          while (realElem.$.classList.length) {
//                //            realElem.$.classList.remove(realElem.$.classList[0]);
//                //          }
//                //        delete cssAnimationObject[realElem.$.id];
//                //        realElem.$.removeAttribute("data-animation");
//                //        elem.dataset.ckeRealelement = encodeURIComponent(realElem.$.outerHTML);
//                //        elem.classList.remove(animClassPrefix+elem.id);
//            } else {
//          if (elem.dataset.ckeRealElementType === undefined) {
            var i;
            for (i=elem.$.classList.length - 1; i>=0; i-=1) {
                if (elem.$.classList[i].indexOf(animClassPrefix) >= 0) {
                    elem.$.classList.remove(elem.$.classList[i]);
                }
            }
//            delete cssAnimationObject[elem.getId()];
            elem.removeAttribute("data-animation");
//            }
        },
//        getSelectedElem: function () {
//        },
        // Get list of TAGS from element to top.
        getPathToTop: function (element) {
            var nodes = [];
            var elem = element;
            var cnt = 0;
            while (elem) {
                if (elem.getName() === "body") {
                    break;
                }
                // Skip Fake Elements !!
                if (elem.$.dataset.ckeRealElementType === undefined) {
                	nodes[cnt] = elem;
	                cnt += 1;
                }
                elem = elem.getParent();
            }
            return nodes;
        },
        // Called when clicking the "test It" button in the add animation dialog
        cssanimAddAnimDialogTest: function (btn) {
            var testDiv = this.cssanimAddAnimDialog.getElementsByClassName('cke_cssanim_container_div')[0];
            var tab, sel, inp;
            var animName, animTiming, animDir, animDuration, animDelay, animIter;
            if (btn.name === "overBtn") {
                tab = this.cssanimAddAnimDialog.querySelector('#cssanimAddAnimDialogTabOver');
                sel = tab.getElementsByTagName('select');
                inp = tab.getElementsByTagName('input');
//                animName = myGetNamedItem(sel,'anim_O').selectedOptions[0].value;
//                animTiming = myGetNamedItem(sel,'timing_O').selectedOptions[0].value;
//                animDir = myGetNamedItem(sel,'direction_O').selectedOptions[0].value;
                animName = myGetNamedItem(sel,'anim_O').value;
                animTiming = myGetNamedItem(sel,'timing_O').value;
                animDir = myGetNamedItem(sel,'direction_O').value;
                animDuration = myGetNamedItem(inp,'duration_O').value + "s";
                animDelay = myGetNamedItem(inp,'delay_O').value + "s";
                animIter = myGetNamedItem(inp,'iteration_O').value;
                if (animIter === "0") {
                    animIter = "infinite";
                }
                if (animName !== "none") {
                	testDiv.style.animation = "none";
                	testDiv.style.webkitAnimation = "none";
                    setTimeout(function() {
                        testDiv.style.animation = animName + " " + animDuration + " " + animTiming + " " + animDelay + " " + animIter + " " + animDir;
                        testDiv.style.webkitAnimation =  animName + " " + animDuration + " " + animTiming + " " + animDelay + " " + animIter + " " + animDir;
                   }, 10);
                } else {
                    testDiv.style.animation = "none";
                	testDiv.style.webkitAnimation = "none";
                }
            }
            if (btn.name === "clickBtn") {
                tab = this.cssanimAddAnimDialog.querySelector('#cssanimAddAnimDialogTabClick');
                sel = tab.getElementsByTagName('select');
                inp = tab.getElementsByTagName('input');
//                animName = myGetNamedItem(sel,'anim_C').selectedOptions[0].value;
//                animTiming = myGetNamedItem(sel,'timing_C').selectedOptions[0].value;
//                animDir = myGetNamedItem(sel,'direction_C').selectedOptions[0].value;
                animName = myGetNamedItem(sel,'anim_C').value;
                animTiming = myGetNamedItem(sel,'timing_C').value;
                animDir = myGetNamedItem(sel,'direction_C').value;
                animDuration = myGetNamedItem(inp,'duration_C').value + "s";
                animDelay = myGetNamedItem(inp,'delay_C').value + "s";
                animIter = myGetNamedItem(inp,'iteration_C').value;
                if (animIter === "0") {
                    animIter = "infinite";
                }
                if (animName !== "none") {
                	testDiv.style.animation = "none";
                	testDiv.style.webkitAnimation = "none";
                  setTimeout(function() {
                        testDiv.style.animation = animName + " " + animDuration + " " + animTiming + " " + animDelay + " " + animIter + " " + animDir;
                        testDiv.style.webkitAnimation =  animName + " " + animDuration + " " + animTiming + " " + animDelay + " " + animIter + " " + animDir;
                   }, 10);
                } else {
                    testDiv.style.animation = "none";
                	testDiv.style.webkitAnimation = "none";
               }
            }
            if (btn.name === "loadBtn") {
                tab = this.cssanimAddAnimDialog.querySelector('#cssanimAddAnimDialogTabLoad');
                sel = tab.getElementsByTagName('select');
                inp = tab.getElementsByTagName('input');
//                animName = myGetNamedItem(sel,'anim_L').selectedOptions[0].value;
//                animTiming = myGetNamedItem(sel,'timing_L').selectedOptions[0].value;
//                animDir = myGetNamedItem(sel,'direction_L').selectedOptions[0].value;
                animName = myGetNamedItem(sel,'anim_L').value;
                animTiming = myGetNamedItem(sel,'timing_L').value;
                animDir = myGetNamedItem(sel,'direction_L').value;
                animDuration = myGetNamedItem(inp,'duration_L').value + "s";
                animDelay = myGetNamedItem(inp,'delay_L').value + "s";
                animIter = myGetNamedItem(inp,'iteration_L').value;
                if (animIter === "0") {
                    animIter = "infinite";
                }
                if (animName !== "none") {
                    // "bounceOut 3s linear 0s 1 normal";
                    testDiv.addEventListener('animationend', function () {
                        this.style.animation = "none";
                    }, false);
                	testDiv.style.animation = "none";
                	testDiv.style.webkitAnimation = "none";
                    setTimeout(function() {
                        testDiv.style.animation = animName + " " + animDuration + " " + animTiming + " " + animDelay + " " + animIter + " " + animDir;
                        testDiv.style.webkitAnimation =  animName + " " + animDuration + " " + animTiming + " " + animDelay + " " + animIter + " " + animDir;
                    }, 10);
                } else {
                    testDiv.style.animation = "none";
                	testDiv.style.webkitAnimation = "none";
               }
            }
        },
        getAvailableAnims: function () {
            var res = {};
            res.allAvailableAnimsCssAnim = this.allAvailableAnimsCssAnim;
            res.allAvailableAnimsCustom = this.allAvailableAnimsCustom;
            res.allAvailableAnimsConflict = this.allAvailableAnimsConflict;
            return res;
        },
        // load a custom css file // Asynchronous process !!!
        getCustomCss: function (obj) {
            var divRes;
            var cssFile;
            function addCustomCssFunc(cssFile, css, obj) {
                // do some cleaning !!
                var candidates = document.getElementsByTagName('style');
                var i, j, k;
                // need to remove the css element from document if already present !!!!
                for (i = candidates.length - 1; i >= 0; i-=1) {
                    if (candidates[i].title === 'cssCustom') {
                        candidates[i].parentNode.removeChild(candidates[i]);
                        break;
                    }
                }
                if (CKEDITOR.plugins.cssanim.ckEditor.config.allowedAnimations.CUSTOM) {
                    delete CKEDITOR.plugins.cssanim.ckEditor.config.allowedAnimations.CUSTOM;
                }
                CKEDITOR.plugins.cssanim.allAvailableAnimsCustom = [];
                CKEDITOR.plugins.cssanim.allAvailableAnimsConflict = [];
                var style = document.createElement('style');
                style.title = "cssCustom";
                style.type = 'text/css';
                style.innerHTML = css;
                document.getElementsByTagName('head')[0].appendChild(style);
                var cssAnims = [];
                var htmlRes = "";
                var liStr;
                var cssAnimStyleSheet = null;
                var lang = CKEDITOR.plugins.cssanim.ckEditor.lang.cssanim;
                // get cssanim.css style sheet
                for (i = document.styleSheets.length - 1; i >= 0; i-=1) {
                    if ((document.styleSheets[i].href) && (document.styleSheets[i].href.indexOf("cssanim.css") >= 0)) {
                        cssAnimStyleSheet = document.styleSheets[i];
                        break;
                    }
                }
                var tmpArr = [];
                for (i = document.styleSheets.length - 1; i >= 0; i-=1) {
                    if (document.styleSheets[i].title === 'cssCustom') {
                        for (j = 0; j < document.styleSheets[i].cssRules.length; j += 1) {
                            if ((document.styleSheets[i].cssRules[j].type === CSSRule.KEYFRAMES_RULE)
                            	|| (document.styleSheets[i].cssRules[j].type === CSSRule.WEBKIT_KEYFRAMES_RULE)){
                            	if (tmpArr.indexOf(document.styleSheets[i].cssRules[j].name) < 0) {
	                                liStr = '<li>' + document.styleSheets[i].cssRules[j].name + '</li>';
	                                cssAnims.push(document.styleSheets[i].cssRules[j].name);
	                                tmpArr.push(document.styleSheets[i].cssRules[j].name);
	                                CKEDITOR.plugins.cssanim.allAvailableAnimsCustom.push(document.styleSheets[i].cssRules[j].name);
	                                if (cssAnimStyleSheet) {
	                                    for (k = 0; k < cssAnimStyleSheet.cssRules.length; k += 1) {
	                                        if (cssAnimStyleSheet.cssRules[k].name === document.styleSheets[i].cssRules[j].name) {
	                                            liStr = '<li><span style="color:tomato">' + document.styleSheets[i].cssRules[j].name + ' '+lang.warningOverride+'</span></li>';
	                                            CKEDITOR.plugins.cssanim.allAvailableAnimsConflict.push(document.styleSheets[i].cssRules[j].name);
	                                            break;
	                                        }
	                                    }
	                                }
	                                htmlRes += liStr;
                            	}
                            }
                        }
                    }
                }
                //console.log("Rules : ", cssAnims);
                if (htmlRes.length > 0) {
                    htmlRes = '<span style="font-weight:bold;">'+lang.animationFound+'</span><br>' + '<ul style="text-align: center; margin:5px; margin-left: 25px;">' + htmlRes + '</ul>';
                    // here we need to add the css in the generated document and in allowedAnims.CUSTOM
                    CKEDITOR.plugins.cssanim.ckEditor.config.allowedAnimations.CUSTOM = cssAnims;
                } else {
                    htmlRes = '<span style="font-weight:bold; color:red;">'+lang.noAnimationFound+'</span><br>';
                    if (CKEDITOR.plugins.cssanim.ckEditor.config.allowedAnimations.CUSTOM) {
                        delete CKEDITOR.plugins.cssanim.ckEditor.config.allowedAnimations.CUSTOM;
                    }
                }
                customCssFile = cssFile.trim();
                CKEDITOR.plugins.cssanim.ckEditor.config.customCssFilePath = cssFile.trim();
                if (obj) {
                    divRes = obj.querySelector("#cssResultsDiv");
                    divRes.innerHTML = htmlRes;
                    divRes.style.display = "block";
                }
            }
            // css file asynchronous load is done
//            function transferCompleteSlow(evt) {
//                setTimeout(function(){ transferComplete(evt); }, 3000);
//            }
            function transferComplete(evt) {
                CKEDITOR.plugins.cssanim.ckEditor.container.setStyle('pointer-events', '');
//                CKEDITOR.plugins.cssanim.ckEditor.container.setStyle('outline', '');
                var www = CKEDITOR.plugins.cssanim.ckEditor.document.getBody();
                www.setStyle('backgroundColor', '');
               var lang = CKEDITOR.plugins.cssanim.ckEditor.lang.cssanim;
               if (evt.target.status === 200) {
                    var css = evt.target.response;
                    addCustomCssFunc(cssFile, css, obj);
                    //      		  var res = checkAnimValidity();
                    //    		  //console.log("checkAnimValidity resArray ----> ", res);
                } else {
                	if (obj && divRes) {
	                    var htmlRes = '<span style="font-weight:bold; color:tomato;">'+ lang.errorLoadingFile + evt.target.responseURL + '</span>';
	                    htmlRes += '<br><span style="font-weight:bold; color:red;">'+lang.loadStatus + evt.target.status + '</span><br>';
	                    divRes.innerHTML = htmlRes;
	                    divRes.style.display = "block";
                	}
                    alert(lang.specFile + evt.target.responseURL + lang.fileNotAvailable);
                }
            }
            // css file asynchronous load failed
           function transferFailed(/*evt*/) {
               var lang = CKEDITOR.plugins.cssanim.ckEditor.lang.cssanim;
               alert(lang.fileCannotAccess);
               CKEDITOR.plugins.cssanim.ckEditor.container.setStyle('pointer-events', '');
               CKEDITOR.plugins.cssanim.ckEditor.container.setStyle('outline', '');
           }
            if (Object.prototype.toString.apply(obj) === "[object String]") {
                cssFile = obj.trim();
                obj = null;
            } else {
                divRes = obj.querySelector("#cssResultsDiv");
                divRes.innerHTML = "";
                divRes.style.display = "none";
                if (divRes.parentElement.elements['cssName'].value !== "undefined") {
                	cssFile = divRes.parentElement.elements['cssName'].value.trim();
                } else {
                	cssFile = "";
                }
            }
            var ext = cssFile.substr(cssFile.lastIndexOf('.') + 1);
            //console.log("ext", ext);
            if (cssFile === this.ckEditor.config.customCssFilePath) {
                if (this.allAvailableAnimsCustom.length > 0) {
                    return;
                }
            }
            if (!cssFile) {
            	// No more custon css file, so remove it from dom and clean related animation from
            	// the list of available animations
                // alert("Removing Custom CSS File !");
                var candidates = document.getElementsByTagName('style');
                var i;
                // need to remove the css element from document if already present !!!!
                for (i = candidates.length - 1; i >= 0; i-=1) {
                    if (candidates[i].title === 'cssCustom') {
                        candidates[i].parentNode.removeChild(candidates[i]);
                        break;
                    }
                }
                if (CKEDITOR.plugins.cssanim.ckEditor.config.allowedAnimations.CUSTOM) {
                    delete CKEDITOR.plugins.cssanim.ckEditor.config.allowedAnimations.CUSTOM;
                }
                customCssFile = "";
                this.ckEditor.config.customCssFilePath = "";
                CKEDITOR.plugins.cssanim.allAvailableAnimsCustom = [];
                CKEDITOR.plugins.cssanim.allAvailableAnimsConflict = [];
                //    		res = checkAnimValidity();
                return;
            }
            var lang = CKEDITOR.plugins.cssanim.ckEditor.lang.cssanim;
           if (ext !== "css") {
                alert(lang.fileSpecified + cssFile + lang.fileExtInvalid);
                return;
            }
            // low connection test
            //cssFile = 'http://deelay.me/1000/'+cssFile;
            this.ckEditor.container.setStyle('pointer-events', 'none');
            var www = this.documentEditor.getBody();
            www.setStyle('backgroundColor', 'Gainsboro ');

            var oReq = new XMLHttpRequest();
            oReq.onload = transferComplete;
            oReq.onerror = transferFailed;
            oReq.open("GET", cssFile, true);
            oReq.send();
        }
    };
})();
