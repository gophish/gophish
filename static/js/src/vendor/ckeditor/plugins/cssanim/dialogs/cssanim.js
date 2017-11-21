/**
 * Copyright (c) 2014, CKSource - Frederico Knabben. All rights reserved.
 * Licensed under the terms of the MIT License (see LICENSE.md).
 *
 * The cssanim plugin dialog window definition.
 *
 * Created out of the CKEditor Plugin SDK:
 * http://docs.ckeditor.com/#!/guide/plugin_sdk_sample_1
 */
// css = cssSelectorStart
// cds = cssDefStart
// cso = cssSelectorOver
// cdo = cssDefOver
// csc = cssSelectorClick
// cdc = cssDefClick
// ral = css Remove after Load flag
// Our dialog definition.
CKEDITOR.dialog.add('cssanimMainDialog', function (editor) {
    "use strict";
    var lang = editor.lang.cssanim;

    function rgbToHex(color) {
        if (color.substr(0, 1) === "#") {
            return color;
        }
        var nums = /(.*?)rgb\((\d+),\s*(\d+),\s*(\d+)\)/i.exec(color),
            r = parseInt(nums[2], 10).toString(16),
            g = parseInt(nums[3], 10).toString(16),
            b = parseInt(nums[4], 10).toString(16);
        return "#" + (
            (r.length === 1 ? "0" + r : r) + (g.length === 1 ? "0" + g : g) + (b.length === 1 ? "0" + b : b));
    }

    function getAllowedTagsHtml(tags, onloadTags) {
        var allowedTagsStr = '';
        allowedTagsStr += '<form>';
        allowedTagsStr += '<div style=" overflow: auto; height: 350px; padding-bottom: 10px; margin-bottom: 10px;">';
//        allowedTagsStr += '<table style="width:360px; display:block; overflow-y=scroll; max-height: 300px; overflow-y: auto;">';
        allowedTagsStr += '<table>';
        var i;
        var checked;
        for (i = 0; i < onloadTags.length; i += 1) {
            allowedTagsStr += '<tr>';
            allowedTagsStr += '<td style="width:280px;">';
            allowedTagsStr += onloadTags[i].toUpperCase();
            allowedTagsStr += '</td>';
            allowedTagsStr += '<td style="text-align:center; width:80px;">';
            checked = "";
            if (tags.indexOf(onloadTags[i]) >= 0) {
                checked = "checked=checked";
            }
            allowedTagsStr += ' <input type="checkbox" name="tags" ' + checked + ' value="' + onloadTags[i] + '">';
            allowedTagsStr += '</td>';
            allowedTagsStr += '</tr>';
        }
        allowedTagsStr += '</table>';
        allowedTagsStr += '</div>';
       allowedTagsStr += '</form>';
        return allowedTagsStr;
    }

    function getparametersHtml() {
//        var lang = editor.lang.cssanim;
        var parametersStr = '';
        var rgbHex = editor.config.highlightBGColor;
        if (editor.config.highlightBGColor.indexOf('rgb') >= 0) {
            rgbHex = rgbToHex(editor.config.highlightBGColor);
        }
        var customCssFilePath = editor.config.customCssFilePath;
        parametersStr += '<form>';
        parametersStr += '<fieldset style="border-radius:10px">';
        parametersStr += '<legend style="font-weight:bold;">'+lang.HBS+'</legend>';
        parametersStr += '<table style="width:300px;">';
        parametersStr += '<tr>';
        parametersStr += '<td style="width:250px;">';
        parametersStr += lang.HBGC;
        parametersStr += '</td>';
        parametersStr += '<td>';
        parametersStr += '<input style="width:50px; height:25px; " type="color" name="HLBgColor" value="' + rgbHex + '"  onchange="CKEDITOR.plugins.cssanim.setHighLightBgColor(this.value)">';
        parametersStr += '</td>';
        parametersStr += '</tr>';
        parametersStr += '<tr>';
        parametersStr += '<td>';
        parametersStr += lang.HBW;
        parametersStr += '</td>';
        parametersStr += '<td style="padding:2px;">';
        parametersStr += '<input style="width:40px; border: 1px solid #ccc;" name="width" type="number" min="1" max="10"  onchange="CKEDITOR.plugins.cssanim.setHighLightBorder(this.value)"  value="' + parseInt(editor.config.highlightBorder) + '"> px';
        parametersStr += '</td>';
        parametersStr += '</tr>';
        parametersStr += '</table>';
        parametersStr += '</fieldset>';
        parametersStr += '<br>';
        parametersStr += '<fieldset style="border-radius:10px">';
        parametersStr += '<legend style="font-weight:bold;">'+lang.CustomCSS+'</legend>';
        parametersStr += '<table style="width:300px;">';
        parametersStr += '<tr>';
        parametersStr += '<td style="width:50px;">';
        parametersStr += lang.cssPath + '&nbsp;';
        parametersStr += '</td>';
        parametersStr += '<td style="width:250px;">';
        parametersStr += '<input type="text" style="width:400px; border: 1px solid #ccc;" name="cssName" value=' + customCssFilePath + '>';
        parametersStr += '</td>';
        parametersStr += '</tr>';
        parametersStr += '<tr>';
        parametersStr += '<td colspan="2" style="text-align:center; padding-top: 8px;color: tomato;">';
        parametersStr += lang.cssInfo;
        parametersStr += '</td>';
        parametersStr += '</tr>';
        parametersStr += '</table>';
        parametersStr += '</fieldset>';
        parametersStr += '<div id="cssResultsDiv" style="margin-top:10px; padding:5px; text-align: center; border: #ccc solid 1px; display:none;">';
        parametersStr += 'test';
        parametersStr += '</div>';
        parametersStr += '</form>';
        return parametersStr;
    }

    function getTabDocAnimationHtml(myDialog) {
//       var lang = editor.lang.cssanim;
       //console.log("availableAnimsObj", myDialog.availableAnimsObj);
        //   	ADD HERE COLOR FOR ANIMATIONS DEPENDING OF CSS ORIGIN :
        //		lightblue if found in cssanim !!
        //		green if found in custom !!
        //		red if not found in any !!
        var ral = lang.irrelevant;
        var inputs = myDialog.getContentElement('tab-parameters-box', 'parametersSettings').getElement().getElementsByTag('input');
        var HLBgColor;
        var input, i;
        for (i = 0; i < inputs.count(); i += 1) {
            input = inputs.getItem(i).getNameAtt();
            if (input === 'HLBgColor') {
                HLBgColor = inputs.getItem(i).getValue();
                break;
            }
        }
        var thStr = '<th style="border: 1px solid gray; text-align:center; padding:5px; font-weight:bold; background-color: lightgray;">';
        var tdStrS = '<td style="border: 1px solid gray; text-align:center; padding:5px;';
        var tdStrE;
        var elems = CKEDITOR.plugins.cssanim.getAllAnimElem(editor);
        var html = lang.noElements;
        var htmlTable = '';
        //     	htmlTable += '<p style="text-align: center; padding: 5px; color: grey; font-weight:bold;">Animation Colors Legend';
        //    	htmlTable += '</p>';
        htmlTable += '<div style="width: 90%; height: 1px; background: grey; margin-left: auto;margin-right: auto;"></div>';
        htmlTable += '<p style="text-align: center; padding: 5px; color: grey; font-weight:bold;">';
        htmlTable += '<span style="color: LightSkyBlue; font-weight:bold;">'+lang.animFromDef+'</span>';
        htmlTable += '<br><span style="color: cornflowerblue; font-weight:bold;">'+lang.animPause+'</span>';
        htmlTable += '<br><span style="color: PaleGreen; font-weight:bold;">'+lang.animFromCustom+'</span>';
        htmlTable += '<br><span style="color: orange; font-weight:bold;">'+lang.animBoth+'</span>';
        htmlTable += '<br><span style="color: Red; font-weight:bold;">'+lang.animUndef+'</span>';
        htmlTable += '<br>';
        htmlTable += '</p>';
        htmlTable += '<div style="width: 90%; height: 1px; background: grey; margin-left: auto;margin-right: auto;"></div>';
        htmlTable += '<p style="text-align: center; padding: 5px; color: grey; font-weight:bold;">';
        htmlTable += lang.dblClick;
        htmlTable += '</p>';
        htmlTable += '<div style="width: 90%; height: 1px; background: grey; margin-left: auto;margin-right: auto;margin-bottom: 10px;"></div>';
        htmlTable += '<div style=" width: 600px; overflow: auto; height: 400px;">';
        //htmlTable += '<table style=" display: inline-block; height: 300px;">';
        htmlTable += '<table>';
        htmlTable += '<tr>';
        htmlTable += thStr + 'Tag</th>';
        htmlTable += thStr + 'Id</th>';
        htmlTable += thStr + 'OnLoad</th>';
        htmlTable += thStr + 'OnClick</th>';
        htmlTable += thStr + 'OnOver</th>';
        htmlTable += thStr + lang.ral +'</th>';
        htmlTable += '</tr>';
        var cnt = 0;
        var bgColor = 'azure;';
        var tag;
        var exp, animObj;
        var curElem;
        var key;
        for (key in elems) {
	        if (elems.hasOwnProperty(key)) {
	        	curElem = elems[key];
	        	ral = lang.irrelevant;
	            tdStrE = '">';
	            cnt+=1;
	            bgColor = 'azure';
	            tag = curElem.elem.nodeName.toLowerCase();
	            if (editor.config.allowedTags.indexOf(tag) < 0) {
	                bgColor = 'MistyRose';
	            }
	            htmlTable += '<tr style="background-color:' + bgColor + ';" ';
	            htmlTable += ' onmouseover="this.style.backgroundColor=\'' + HLBgColor + '\'; CKEDITOR.plugins.cssanim.runHighLightElemById(\'' + curElem.elem.id + '\'); "';
	            htmlTable += ' onmouseout="this.style.backgroundColor=\'' + bgColor + '\'; CKEDITOR.plugins.cssanim.cleanHighlight();"';
	            htmlTable += ' ondblclick=" CKEDITOR.plugins.cssanim.runAddAnimDialogById(\'' + curElem.elem.id + '\')"';
	            htmlTable += ' >';
	            htmlTable += tdStrS + tdStrE;
	            htmlTable += curElem.elem.nodeName;
	            htmlTable += '</td>';
	            htmlTable += tdStrS + tdStrE;
	            htmlTable += curElem.elem.id;
	            htmlTable += '</td>';
	            tdStrE = '">';
	            if (curElem.cds) {
	                exp = curElem.cds.replace('{animation:', '').replace('}', '').replace(';', '').trim();
	                animObj = parseCSS3AnimationShorthand(exp);
	                ral = (curElem.ral) ? "YES" : "NO";
	                if (animObj.name) {
	                	if (animObj.name === 'cssAnimPause') {
	                        tdStrE = ' background-color:cornflowerblue;">';              		
	                	} else if (myDialog.availableAnimsObj && myDialog.availableAnimsObj.allAvailableAnimsConflict && (myDialog.availableAnimsObj.allAvailableAnimsConflict.indexOf(animObj.name) >= 0)) {
	                        tdStrE = ' background-color:orange;">';
	                    } else if (myDialog.availableAnimsObj && myDialog.availableAnimsObj.allAvailableAnimsCustom && (myDialog.availableAnimsObj.allAvailableAnimsCustom.indexOf(animObj.name) >= 0)) {
	                        tdStrE = ' background-color:PaleGreen;">';
	                    } else if (myDialog.availableAnimsObj && myDialog.availableAnimsObj.allAvailableAnimsCssAnim && (myDialog.availableAnimsObj.allAvailableAnimsCssAnim.indexOf(animObj.name) >= 0)) {
	                        tdStrE = ' background-color:LightSkyBlue;">';
	                    } else {
	                        tdStrE = ' background-color:Red;">';
	                    }
	                }
	                htmlTable += tdStrS + tdStrE;
	            } else {
	                animObj = {};
	                animObj.name = "";
	                htmlTable += tdStrS + tdStrE;
	            }
	            htmlTable += animObj.name;
	            htmlTable += '</td>';
	            tdStrE = '">';
	            if (curElem.cdc) {
	            	if (curElem.cdc.indexOf('animation-play-state') >= 0) {
	            		animObj = {
	            	            name: 'cssAnimPause',
	            	            duration: null,
	            	            timingFunction: 'linear',
	            	            delay: 0,
	            	            iterationCount: 1,
	            	            direction: 'normal'
	            	        };
	            	} else {
	    	            exp = curElem.cdc.replace('{animation:', '').replace('}', '').replace(';', '').trim();
	    	            animObj = parseCSS3AnimationShorthand(exp);    		
	            	}
	                if (animObj.name) {
	                	if (animObj.name === 'cssAnimPause') {
	                        tdStrE = ' background-color:cornflowerblue;">';              		
	                	} else if (myDialog.availableAnimsObj && myDialog.availableAnimsObj.allAvailableAnimsConflict && (myDialog.availableAnimsObj.allAvailableAnimsConflict.indexOf(animObj.name) >= 0)) {
	                        tdStrE = ' background-color:orange;">';
	                    } else if (myDialog.availableAnimsObj && myDialog.availableAnimsObj.allAvailableAnimsCustom && (myDialog.availableAnimsObj.allAvailableAnimsCustom.indexOf(animObj.name) >= 0)) {
	                        tdStrE = ' background-color:PaleGreen;">';
	                    } else if (myDialog.availableAnimsObj && myDialog.availableAnimsObj.allAvailableAnimsCssAnim && (myDialog.availableAnimsObj.allAvailableAnimsCssAnim.indexOf(animObj.name) >= 0)) {
	                        tdStrE = ' background-color:LightSkyBlue;">';
	                    } else {
	                        tdStrE = ' background-color:Red;">';
	                    }
	                }
	                htmlTable += tdStrS + tdStrE;
	            } else {
	                animObj = {};
	                animObj.name = "";
	                htmlTable += tdStrS + tdStrE;
	            }
	            htmlTable += animObj.name;
	            htmlTable += '</td>';
	            tdStrE = '">';
	            if (curElem.cdo) {
	            	if (curElem.cdo.indexOf('animation-play-state') >= 0) {
	            		animObj = {
	            	            name: 'cssAnimPause',
	            	            duration: null,
	            	            timingFunction: 'linear',
	            	            delay: 0,
	            	            iterationCount: 1,
	            	            direction: 'normal'
	            	        };
	            	} else {
	    	            exp = curElem.cdo.replace('{animation:', '').replace('}', '').replace(';', '').trim();
	    	            animObj = parseCSS3AnimationShorthand(exp);    		
	            	}
	                if (animObj.name) {
	                	if (animObj.name === 'cssAnimPause') {
	                        tdStrE = ' background-color:cornflowerblue;">';              		
	                	} else if (myDialog.availableAnimsObj && myDialog.availableAnimsObj.allAvailableAnimsConflict && (myDialog.availableAnimsObj.allAvailableAnimsConflict.indexOf(animObj.name) >= 0)) {
	                        tdStrE = ' background-color:orange;">';
	                    } else if (myDialog.availableAnimsObj && myDialog.availableAnimsObj.allAvailableAnimsCustom && (myDialog.availableAnimsObj.allAvailableAnimsCustom.indexOf(animObj.name) >= 0)) {
	                        tdStrE = ' background-color:PaleGreen;">';
	                    } else if (myDialog.availableAnimsObj && myDialog.availableAnimsObj.allAvailableAnimsCssAnim && (myDialog.availableAnimsObj.allAvailableAnimsCssAnim.indexOf(animObj.name) >= 0)) {
	                        tdStrE = ' background-color:LightSkyBlue;">';
	                    } else {
	                        tdStrE = ' background-color:Red;">';
	                    }
	                }
	                htmlTable += tdStrS + tdStrE;
	            } else {
	                animObj = {};
	                animObj.name = "";
	                htmlTable += tdStrS + tdStrE;
	            }
	            htmlTable += animObj.name;
	            htmlTable += '</td>';
	            tdStrE = '">';
	            htmlTable += tdStrS + tdStrE;
	            htmlTable += ral;
	            htmlTable += '</td>';
	            htmlTable += '</tr>';
	        }
        }
        if (cnt > 0) {
            htmlTable += '</table>';
            htmlTable += '</div>';
            htmlTable += '<div style="height: 6px;"></div>';
            html = htmlTable;
        }
        return html;
    }

    return {
        // Basic properties of the dialog window: title, minimum size.
        title: editor.lang.cssanim.genProperties,
        minWidth: 550,
        width: 550,
        minHeight: 250,
        height: 250,
        resizable: CKEDITOR.DIALOG_RESIZE_NONE,
        // Dialog window content definition.
        contents: [
            // Definition of the parameters Box Settings dialog tab (page).
            {
                id: 'tab-doc-animations',
                label: lang.docAnim,
                elements: [{
                    type: 'html',
                    id: "allAnimations",
                    html: ''
//                    setup: function () {},
//                    commit: function () {}
                }]
            }, {
                // Definition of the Allowed Tags Settings dialog tab (page).
                id: 'tab-allowed-tags',
                label: lang.allowedTags,
                // The tab content.
                elements: [{
                    type: 'html',
                    id: "allowedTags",
                    allowedTags: null,
                    html: '',
                    setup: function () {
                        this.allowedTags = editor.config.allowedTags;
                    },
                    cancel: function () {
                        editor.config.allowedTags = this.allowedTags;
                        var inputs = this.getElement().getElementsByTag('input');
                        var i;
                        for (i = 0; i < inputs.count(); i += 1) {
                            if (editor.config.allowedTags.indexOf(inputs.getItem(i).$.value) < 0) {
                                inputs.getItem(i).$.checked = false;
                            } else {
                                inputs.getItem(i).$.checked = true;
                            }
                        }
                    },
                    commit: function () {
                        var inputs = this.getElement().getElementsByTag('input');
                        var i;
                        editor.config.allowedTags = [];
                        for (i = 0; i < inputs.count(); i += 1) {
                            if (inputs.getItem(i).$.checked) {
                                editor.config.allowedTags.push(inputs.getItem(i).$.value);
                            }
                        }
                    }
                }]
            }, {
                id: 'tab-parameters-box',
                label: lang.advSettings,
                // Require the id attribute to be enabled.
                elements: [{
                    type: 'html',
                    id: "parametersSettings",
                    html: '',
                    bg: '',
                    bw: 0,
                    pad: 0,
                    setup: function () {
                        var inputs = this.getElement().getElementsByTag('input');
                        var input, i;
                        for (i = 0; i < inputs.count(); i += 1) {
                            input = inputs.getItem(i).$.name;
                            if (input === 'HLBgColor') {
                                this.bg = inputs.getItem(i).$.value;
                            } else if (input === 'width') {
                                this.bw = inputs.getItem(i).$.value;
                            }
                        }
                        //console.log(" ----------------------- parametersSettings setup -----------", this);
                    },
                    commit: function () {
                        var inputs = this.getElement().getElementsByTag('input');
                        var input, i;
                        for (i = 0; i < inputs.count(); i += 1) {
                            input = inputs.getItem(i).$.name;
                            if (input === 'HLBgColor') {
                                editor.config.highlightBGColor = inputs.getItem(i).$.value;
                            } else if (input === 'width') {
                                editor.config.highlightBorder = inputs.getItem(i).$.value + "px";
//                            } else if (input === 'padding') {
//                                editor.config.highlightPadding = inputs.getItem(i).$.value + "px";
                            } else if (input === 'cssName') {
                                if (this.getDialog().initialCssFile !== inputs.getItem(i).$.value) {
                                    CKEDITOR.plugins.cssanim.getCustomCss(inputs.getItem(i).$.value);
                                }
                            }
                        }
                        //console.log(" ----------------------- parametersSettings commit -----------", inputs);
                    },
                    cancel: function () {
                        var inputs = this.getElement().getElementsByTag('input');
                        var input, i;
                        for (i = 0; i < inputs.count(); i += 1) {
                            input = inputs.getItem(i).$.name;
                            if (input === 'HLBgColor') {
                                inputs.getItem(i).$.value = this.bg;
                                editor.config.highlightBGColor = this.bg;
                            } else if (input === 'width') {
                                inputs.getItem(i).$.value = this.bw;
                                editor.config.highlightBorder = this.bw + "px";
                            }
                        }
                        //console.log(" ----------------------- parametersSettings cancel -----------", this);
                    }
                }]
            }
        ],
        onLoad: function () {
            CKEDITOR.plugins.cssanim.setCssAnimDialog(this);
            this.initialCssFile = "";
            // Init dialog tabs
            var myDomId, myEl;
            myDomId = this.getContentElement('tab-doc-animations', 'allAnimations').domId;
            myEl = document.getElementById(myDomId);
            myEl.innerHTML = getTabDocAnimationHtml(this);
            myDomId = this.getContentElement('tab-allowed-tags', 'allowedTags').domId;
            myEl = document.getElementById(myDomId);
            myEl.innerHTML = getAllowedTagsHtml(editor.config.allowedTags, editor.config.onLoadAllowedTags);
            myDomId = this.getContentElement('tab-parameters-box', 'parametersSettings').domId;
            myEl = document.getElementById(myDomId);
            myEl.innerHTML = getparametersHtml();
            this.on('selectPage', function (e) {
                //console.log("SELECT PAGE ------------------------------------------", e.data.page);
                var domId, el;
                var form;
                if (e.data.currentPage === "tab-allowed-tags") {
                    domId = e.sender.getContentElement('tab-allowed-tags', 'allowedTags');
                    domId.commit();
                }
                if (e.data.page === "tab-doc-animations") {
                    domId = e.sender.getContentElement('tab-doc-animations', 'allAnimations').domId;
                    el = document.getElementById(domId);
                    el.innerHTML = getTabDocAnimationHtml(this);
                } else if (e.data.page === "tab-allowed-tags") {
                    domId = e.sender.getContentElement('tab-allowed-tags', 'allowedTags').domId;
                    el = document.getElementById(domId);
                    el.innerHTML = getAllowedTagsHtml(editor.config.allowedTags, editor.config.onLoadAllowedTags);
                } else if (e.data.page === "tab-parameters-box") {
                    form = e.sender.getContentElement('tab-parameters-box', 'parametersSettings');
                    domId = form.domId;
                    el = document.getElementById(domId);
                    el.innerHTML = getparametersHtml();
                }
            });
        },
        onShow: function () {
            //console.log("ON SHOW CSSANIM ...............");
            this.setupContent();
            var form = this.getContentElement('tab-parameters-box', 'parametersSettings');
            var domId = form.domId;
            var el = document.getElementById(domId);
            if (editor.config.customCssFilePath !== this.initialCssFile) {
                CKEDITOR.plugins.cssanim.getCustomCss(el);
                this.initialCssFile = editor.config.customCssFilePath;
            }
            this.availableAnimsObj = CKEDITOR.plugins.cssanim.getAvailableAnims();
            // this 3 following lines are a trick to force refresh the tab !!
            this.hidePage('tab-doc-animations');
            this.showPage('tab-doc-animations');
            this.selectPage('tab-doc-animations');
            CKEDITOR.plugins.cssanim.cleanHighlight();
        },
        // This method is invoked once a user clicks the OK button, confirming the dialog.
        onOk: function () {
            //console.log("ON OK CSSANIM ...............");
            //			// Invoke the commit methods of all dialog window elements, so the <cssanimMainDialog> element gets modified.
            this.commitContent();
            CKEDITOR.plugins.cssanim.pendingChanges = [];
        },
//        onHide: function () {
//            //console.log("ON HIDE CSSANIM ...............");
//        },
        onCancel: function () {
            //console.log("ON CANCEL CSSANIM ...............");
            if (CKEDITOR.plugins.cssanim.managePending) {
                // need to cancel changes on pending elements
                //console.log("ON CANCEL CSSANIM  PENDING ELEMENTS ============== ", CKEDITOR.plugins.cssanim.pendingChanges);
                var i, obj;
                //				for(i = 0; i < CKEDITOR.plugins.cssanim.pendingChanges.length; i+=1 ) {
                for (i = CKEDITOR.plugins.cssanim.pendingChanges.length - 1; i >= 0; i -= 1) {
                    //console.log("To Be Canceled:", CKEDITOR.plugins.cssanim.pendingChanges[i][0], CKEDITOR.plugins.cssanim.pendingChanges[i][1]);
                    obj = JSON.parse(CKEDITOR.plugins.cssanim.pendingChanges[i][1]);
                    //console.log("To Be Canceled:", obj);
                    CKEDITOR.plugins.cssanim.restoreAnimOnElemById(CKEDITOR.plugins.cssanim.pendingChanges[i][0], obj);
                }
                CKEDITOR.plugins.cssanim.pendingChanges = [];
            }
            // restore initial css file
            if (editor.config.customCssFilePath !== this.initialCssFile) {
                var form = this.getContentElement('tab-parameters-box', 'parametersSettings');
                var el = document.getElementById(form.domId);
                var divRes = el.querySelector("#cssResultsDiv");
                divRes.parentElement.elements["cssName"].value = this.initialCssFile.trim();
                CKEDITOR.plugins.cssanim.getCustomCss(el);
            }
            this.foreach(function (widget) {
//                // Make sure IE triggers "change" event on last focused input before closing the dialog. (#7915)
//                if (CKEDITOR.env.ie && (this._.currentFocusIndex === widget.focusIndex)) {
//                	widget.getInputElement().$.blur();
//                }
                if (widget.cancel) {
                	widget.cancel.apply(widget);
                }
            });
        }
    };
});