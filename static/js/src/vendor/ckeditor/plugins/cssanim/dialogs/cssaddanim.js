/**
 * @license Copyright (c) 2003-2015, CKSource - Frederico Knabben. All rights reserved.
 * For licensing, see LICENSE.md or http://ckeditor.com/license
 */
"use strict";
var selectList_cache;
function iterationChange(obj) {
    var myDoc = obj.ownerDocument;
    var mySpan = myDoc.getElementById('rmAfterLoadSpan');
    var myInput = myDoc.getElementById('rmAfterLoadInput');
    if (obj.value === "0") {
        mySpan.style.color = "lightgrey";
        myInput.disabled = true;
    } else {
        mySpan.style.color = "black";
        myInput.disabled = false;
    }
}
function animSelectionChange(obj) {
	//console.log('animSelectionChange', obj);
	var lang = CKEDITOR.plugins.cssanim.ckEditor.lang.cssanim;
	var myTr = obj.parentElement.parentElement;
	var myBtn = myTr.parentNode.parentNode.parentNode.getElementsByClassName('cssAnimButton');
    var myDoc = obj.ownerDocument;
    var mySpan = myDoc.getElementById('rmAfterLoadSpan');
    var myInput = myDoc.getElementById('rmAfterLoadInput');
    var anim_l_obj = myDoc.getElementsByName('anim_L').item(0);
    var anim_c_obj = myDoc.getElementsByName('anim_C').item(0);
    var anim_o_obj = myDoc.getElementsByName('anim_O').item(0);
    //console.log(anim_l_obj, anim_c_obj, anim_o_obj, selectList_cache);
	var i;
	if ((obj.name === "anim_L") && (anim_l_obj.value === "none")) {
		if (anim_c_obj.value === "cssAnimPause") {
			alert(lang.pbOnClickPause);
			anim_l_obj.selectedIndex = selectList_cache;
			return;
		}
		if (anim_o_obj.value === "cssAnimPause") {
			alert(lang.pbOnOverPause);
			anim_l_obj.selectedIndex = selectList_cache;
			return;
		}
	}
	if (obj.name === "anim_C") {
		if ((anim_l_obj.value === "none") && (anim_c_obj.value === "cssAnimPause")) {
			alert(lang.pbOnClickOnLoadNone);
			anim_c_obj.selectedIndex = selectList_cache;
			return;
		}
	}
	if (obj.name === "anim_O") {
		if ((anim_l_obj.value === "none") && (anim_o_obj.value === "cssAnimPause")) {
			alert(lang.pbOnOverOnLoadNone);
			anim_o_obj.selectedIndex = selectList_cache;
			return;
		}
	}
	if ((obj.value === 'cssAnimPause') || (obj.value === 'none')) {
		for (i=1; i<myTr.childElementCount; i+=1) {
			myTr.childNodes[i].style.opacity = 0;
		}
		myBtn[0].style.display = 'none';
		if (obj.name === "anim_L") {
	        mySpan.style.color = "lightgrey";
	        myInput.disabled = true;
		}
	} else {
		for (i=1; i<myTr.childElementCount; i+=1) {
			myTr.childNodes[i].style.opacity = 'initial';
		}
		myBtn[0].style.display = 'initial';
		if (obj.name === "anim_L") {
			mySpan.style.color = "black";
			myInput.disabled = false;
		}
	}
}

CKEDITOR.dialog.add('cssanimAddAnimDialog', function (editor) {
    var lang = editor.lang.cssanim;
    var imagePath = CKEDITOR.getUrl(CKEDITOR.plugins.get('cssanim').path + 'dialogs/css3animation.png');
    var attachedElem = null;
    var pendingChangesObj;

    function getAnimsHtml(allowedAnims, suffix, animation) {
//        var lang = editor.lang.cssanim;
        var allowedAnimsStr = "";
        var selected;
        var i;
        allowedAnimsStr += "<select  onclick=\"selectList_cache=this.selectedIndex\" onChange=\"animSelectionChange(this);\" name=\"anim" + suffix + "\">";
        selected = "";
        if (animation === "none") {
        	selected = "selected";
        }
        allowedAnimsStr += '<option value="none" '+selected+'>'+lang.none+'<\/option>';
        if (suffix !== '_L') {
	        // add specific "pause" animation
	        selected = "";
	        if (animation === "cssAnimPause") {
	        	selected = "selected";
	        }
        allowedAnimsStr += '<option value="cssAnimPause" '+selected+'>'+lang.cssAnimPause+'<\/option>';
        }
        var key, tabAnims;
        for (key in allowedAnims) {
            tabAnims = allowedAnims[key];
            allowedAnimsStr += "<optgroup label=\"" + key + "\" style=\"font-weight:bold\">";
            for (i = 0; i < tabAnims.length; i += 1) {
                selected = "";
                if (tabAnims[i] === animation) {
                    selected = "selected";
                }
                allowedAnimsStr += '<option value="' + tabAnims[i] + '" ' + selected + '>' + tabAnims[i] + '<\/option>';
            }
            allowedAnimsStr += "  <\/optgroup>";
        }
        allowedAnimsStr += "<\/select>";
        return allowedAnimsStr;
    }

    function getTableHtml(allowedAnimations, suffix, ral, initVal) {
//        var lang = editor.lang.cssanim;
        var tabStr = "";
        var checked = 'checked = "checked"';
        var delay = 0;
        var direction = "normal";
        var duration = 1;
        var iterationCount = 1;
        var animation = "none";
        var timingFunction = "linear";
        var ind;
        var timingValues = ["linear", "ease", "ease-in", "ease-out", "ease-in-out"];
        var dirValues = ["normal", "alternate"];
        var animObj = null;
        var exp;
        var selected = "";
        var tdSpanColor = "black";
//        var tabH = "";
//        if (suffix === '_O') {
//        	tabH = "75px;";
//        } else if (suffix === '_C') {
//        	tabH = "75px;";
//        } else {
//        	tabH = "75px;";
//        }
        if (initVal) {
        	if (initVal.indexOf('animation-play-state') >= 0) {
        		animObj = {
        	            name: 'cssAnimPause',
        	            duration: null,
        	            timingFunction: 'linear',
        	            delay: 0,
        	            iterationCount: 1,
        	            direction: 'normal'
        	        };
        	} else {
	            exp = initVal.replace('{animation:', '').replace('}', '').replace(';', '').trim();
	            animObj = parseCSS3AnimationShorthand(exp);
        	}
            //console.log("initVal animObj ------------->", exp, animObj);
            delay = animObj.delay / 1000;
            direction = animObj.direction;
            duration = animObj.duration / 1000;
            iterationCount = animObj.iterationCount;
            if (iterationCount === "infinite") {
                iterationCount = 0;
                ral = false;
            }
            animation = animObj.name;
            timingFunction = animObj.timingFunction;
        }
        // 75px to be changed depending on button and other !!!!
        tabStr += '<table style="display:block; overflow-y=scroll;  overflow-y: auto; width:550px;">';
        tabStr += '<tr>';
        tabStr += '<th style="text-align: center;font-weight: bold;border: 1px solid grey;padding: 2px; width:100px;">';
        tabStr += lang.name;
        tabStr += '</th>';
        tabStr += '<th style="text-align: center;font-weight: bold;border: 1px solid grey;padding: 2px; width:60px;">';
        tabStr += lang.duration;
        tabStr += '</th>';
        tabStr += '<th style="text-align: center;font-weight: bold;border: 1px solid grey;padding: 2px; width:100px;">';
        tabStr += lang.tFunc;
        tabStr += '</th>';
        tabStr += '<th style="text-align: center;font-weight: bold;border: 1px solid grey;padding: 2px; width:60px;">';
        tabStr += lang.delay;
        tabStr += '</th>';
        tabStr += '<th title="If set to 0, iteration will be set as \'infinite\'" style="text-align: center;font-weight: bold;border: 1px solid grey;padding: 2px; width:70px;">';
        tabStr += lang.iter;
        tabStr += '</th>';
        tabStr += '<th style="text-align: center;font-weight: bold;border: 1px solid grey;padding: 2px;">';
        tabStr += lang.direction;
        tabStr += '</th>';
        tabStr += '</tr>';
        tabStr += '<tr>';
        tabStr += '<td style="text-align: center;padding: 3px;border: 1px solid gainsboro;">';
        tabStr += getAnimsHtml(allowedAnimations, suffix, animation);
        tabStr += '</td>';
        tabStr += '<td style="text-align: center;padding: 3px;border: 1px solid gainsboro;">';
        tabStr += '<input name="duration' + suffix + '" type="number" min="1" max="600" value="' + duration + '" style="width:4em; border:1px solid gainsboro;">sec';
        tabStr += '</td>';
        tabStr += '<td style="text-align: center;padding: 3px;border: 1px solid gainsboro;">';
        tabStr += "<select name=\"timing" + suffix + "\" style=\"border:1px solid gainsboro;\">";
        for (ind = 0; ind < timingValues.length; ind += 1) {
            selected = "";
            if (timingValues[ind] === timingFunction) {
                selected = "selected";
            }
            tabStr += '<option value="' + timingValues[ind] + '" ' + selected + '>' + timingValues[ind] + '<\/option>';
        }
        tabStr += "<\/select>";
        tabStr += '</td>';
        tabStr += '<td style="text-align: center;padding: 3px;border: 1px solid gainsboro;">';
        tabStr += '<input name="delay' + suffix + '" type="number" min="0" max="600" value="' + delay + '" style="width:4em; border:1px solid gainsboro;">sec';
        tabStr += '</td>';
        tabStr += '<td title="If set to 0, iteration will be set as \'infinite\'" style="text-align: center;padding: 3px;border: 1px solid gainsboro;">';
        tabStr += '<input name="iteration' + suffix + '" type="number" min="0" max="99" value="' + iterationCount + '" style="width:3em; border:1px solid gainsboro" onChange="iterationChange(this);">';
        tabStr += '</td>';
        tabStr += '<td style="text-align: center;padding: 3px;border: 1px solid gainsboro;">';
        tabStr += "<select name=\"direction" + suffix + "\" style=\"border:1px solid gainsboro;\">";
        for (ind = 0; ind < dirValues.length; ind += 1) {
            selected = "";
            if (dirValues[ind] === direction) {
                selected = "selected";
            }
            tabStr += '<option value="' + dirValues[ind] + '" ' + selected + '>' + dirValues[ind] + '<\/option>';
        }
        tabStr += "<\/select>";
        tabStr += '</td>';
        tabStr += '</tr>';
        tdSpanColor = "black";
        if (ral !== null) {
            if (ral === false) {
                checked = "";
            }
            if (iterationCount === 0) {
                checked = "disabled";
                tdSpanColor = "lightgrey";
            }
            tabStr += '<tr>';
            tabStr += '<td colspan="3" style="text-align: center;padding: 3px; border-left: 1px solid gainsboro; border-bottom: 1px solid gainsboro;">';
            tabStr += '<span id="rmAfterLoadSpan" style="color:' + tdSpanColor + '" title="'+lang.ralTitle+'">'+lang.ral+'</span>';
            tabStr += '</td>';
            tabStr += '<td colspan="3" style="text-align: left; padding: 3px;  border-right: 1px solid gainsboro; border-bottom: 1px solid gainsboro;">';
            tabStr += '<input  id="rmAfterLoadInput" type="checkbox" name="rmAfterLoad"  ' + checked + '>';
            tabStr += '</td>';
            tabStr += '</tr>';
        }
        tabStr += '</table>';
        return tabStr;
    }

    function getItemByName(elm, name) {
    	var i;
    	for (i=0;i<elm.$.length;i++) {
    		if (elm.$[i].name == name) {
    			return new CKEDITOR.dom.node(elm.$[i]);
    	 	}
    	 }
 		return null;
    }

    function getHtml(elem, allowedAnimations) {
        var obj = null;
        var jsonObj;
//	OLD FAKE ELEMENTS CODE
//        if (elem.dataset.ckeRealElementType !== undefined) {
//            // //     if (false) {
//            //    	  console.log("FAKE !! FAKE !! FAKE !! FAKE !! FAKE !! ");
//            //	        var realElem = CKEDITOR.dom.element.createFromHtml(decodeURIComponent(
//            //	            elem.dataset.ckeRealelement ), editor.document );
//            //	      console.log("realElem", realElem);
//            //	      obj = realElem.$.attributes['data-animation'];
//        } else {
//      if (elem.dataset.ckeRealElementType === undefined) {
//           obj = elem.attributes['data-animation'];
//        }
      obj = elem.getAttribute('data-animation');
        if (obj) {
            //obj = obj.value;
            jsonObj = decodeURIComponent(obj);
            obj = JSON.parse(jsonObj);
            //console.log("ON getHtml cssanimAddAnimDialog", obj);
            if (CKEDITOR.plugins.cssanim.managePending) {
                pendingChangesObj[pendingChangesObj.length] = [elem.getId(), jsonObj];
            }
        }
        var tabStr = "";
        var keyLoad = "OnLoad";
        var keyOver = "OnOver";
        var keyClick = "OnClick";
        var initVal, ral;
        tabStr += "<div class=\"tabs\" style=\"width:600px;\">";
        tabStr += "<div class=\"tab\">";
        tabStr += "<span style=\"font-weight:bold; font-size: larger;\">" + keyOver + "</span>";
        tabStr += "<div id=\"cssanimAddAnimDialogTabOver\" class=\"content\" style=\"height:55px;\">";
        initVal = (obj && obj.cdo) ? obj.cdo : null;
        tabStr += getTableHtml(allowedAnimations, '_O', null, initVal);
        tabStr += '<div style="text-align: center; padding: 2px;">';
        tabStr += "<input type=\"button\" class=\"cssAnimButton\" value=\""+lang.testIt+" !\" name=\"overBtn\" onclick=\"CKEDITOR.plugins.cssanim.cssanimAddAnimDialogTest(this);\">";
        tabStr += "</div> ";
        tabStr += "</div> ";
        tabStr += "<div class=\"tab\">";
        tabStr += "<span style=\"font-weight:bold;font-size: larger;\">" + keyClick + "</span>";
        tabStr += "<div  id=\"cssanimAddAnimDialogTabClick\" class=\"content\" style=\"height:55px;\">";
        initVal = (obj && obj.cdc) ? obj.cdc : null;
        tabStr += getTableHtml(allowedAnimations, '_C', null, initVal);
        tabStr += '<div style="text-align: center; padding: 2px;">';
        tabStr += "<input type=\"button\" class=\"cssAnimButton\" value=\""+lang.testIt+" !\" name=\"clickBtn\" onclick=\"CKEDITOR.plugins.cssanim.cssanimAddAnimDialogTest(this);\">";
        tabStr += "</div> ";
        tabStr += "</div> ";
        tabStr += "<div class=\"tab\">";
        tabStr += "<span style=\"font-weight:bold;font-size: larger;\">" + keyLoad + "</span>";
        tabStr += "<div  id=\"cssanimAddAnimDialogTabLoad\" class=\"content\" style=\"height:85px;\">";
        initVal = (obj && obj.cds) ? obj.cds : null;
        ral = (obj) ? obj.ral : true;
        tabStr += getTableHtml(allowedAnimations, '_L', ral, initVal);
        tabStr += '<div style="text-align: center; padding: 2px;">';
        tabStr += "<input type=\"button\" class=\"cssAnimButton\" value=\""+lang.testIt+" !\" name=\"loadBtn\" onclick=\"CKEDITOR.plugins.cssanim.cssanimAddAnimDialogTest(this);\">";
        tabStr += "</div> ";
        tabStr += "</div> ";
        tabStr += "</div>";
 //       var prefixFree = "<script src=\"./ckeditor/plugins/cssanim/css/prefixfree.min.js\"></script>";
        var html ='<link rel="stylesheet" href="./ckeditor/plugins/cssanim/css/cssanim.css">' + '<style type="text/css">' + '.cke_cssanim_container' + '{' + 'color:#000 !important;' + 'padding:10px 10px 0;' + 'margin-top:5px;' + 'text-align: center;' + '}' + '.cke_cssanim_container p' + '{' + 'margin: 0 0 10px;' + '}' +
            '</style>' + '<div class="cke_cssanim_container">' + '<div class="cke_cssanim_container_div" style="text-align: center; border: 1px solid grey; width:50%;margin-left: auto;margin-right: auto;">' +
            '<img class="cke_cssanim_container_img" src="' + imagePath + '" alt="CSS Animation" style="border: 1px solid grey; padding: 5px; margin: 10px;">' + '</div>' + '<br><br>' + '<p id="elemName" style="font-weight:bold;font-size: larger;text-align: -webkit-center;">' + 'Animation '+lang.forStr+' : ' + attachedElem.getName() + '</p>' + '<br>' + tabStr + '</div>';
 //       html +=  prefixFree;
        return html;
    }
    return {
        title: lang.editAnimsTitle,
        width: 600,
        minWidth: 600,
        height: 400,
        minHeight: 400,
        resizable: CKEDITOR.DIALOG_RESIZE_NONE,
        contents: [{
            id: 'mainPanel',
            label: 'CSS_Animation',
            title: 'Animation',
            //			expand: true,
            padding: 0,
            elements: [{
                type: 'html',
                id: 'htmlAnim',
                html: '',
                setup: function () {
                    //console.log("SETUP htmlAnim");
                    //console.log("SETUP1 allowedTags", editor.config.allowedTags, this.allowedTags);
                    var el = document.getElementById(this.domId);
                    //console.log("cssanimAddAnimDialog setup", el);
                    var allowedAnimations = editor.config.allowedAnimations;
                    CKEDITOR.plugins.cssanim.runHighLightElem(attachedElem);
                    //console.log("cssanimAddAnimDialog allowedAnimations", allowedAnimations);
                    //			        	var html = "<h1> Add Animation on : "+attachedElem.nodeName+"</h1>";
                    el.innerHTML = getHtml(attachedElem, allowedAnimations);
                    CKEDITOR.plugins.cssanim.cssanimAddAnimDialog = el;
                },
                cancel: function () {
                    //console.log("CANCEL htmlAnim");
                    CKEDITOR.plugins.cssanim.cleanHighlight();
                },
                commit: function () {
                    //console.log("COMMIT htmlAnim");
                    CKEDITOR.plugins.cssanim.cleanHighlight();
                    var animation;
                    var inputs = this.getElement().getElementsByTag('input');
                    var selects = this.getElement().getElementsByTag('select');
                    var res = {};
                    res.animStart = null;
                    res.animClick = null;
                    res.animOver = null;
                    res.ral = false;
                    // get OnLoad data
                    animation = getItemByName(selects, 'anim_L').getValue();
                    var animIter;
                    if (animation !== 'none') {
                        res.animStart = animation + " ";
                        res.animStart += getItemByName(inputs, 'duration_L').getValue() + "s ";
                        res.animStart += getItemByName(selects, 'timing_L').getValue() + " ";
                        res.animStart += getItemByName(inputs, 'delay_L').getValue() + "s ";
                        animIter = getItemByName(inputs, 'iteration_L').getValue();
                        if (animIter === "0") {
                            animIter = "infinite";
                        }
                        res.animStart += animIter + " ";
                        res.animStart += getItemByName(selects, 'direction_L').getValue();
                        // Now RAL
                        if ((getItemByName(inputs, 'rmAfterLoad').$.checked) && (animIter !== 'infinite')) {
                        	res.ral = true;
                        }
                    }
                    // get OnClick data
                    animation = getItemByName(selects, 'anim_C').getValue();
                    if (animation !== 'none') {
                        res.animClick = animation + " ";
                        res.animClick += getItemByName(inputs, 'duration_C').getValue() + "s ";
                        res.animClick += getItemByName(selects, 'timing_C').getValue() + " ";
                        res.animClick += getItemByName(inputs, 'delay_C').getValue() + "s ";
                        animIter = getItemByName(inputs, 'iteration_C').getValue();
                        if (animIter === "0") {
                            animIter = "infinite";
                        }
                        res.animClick += animIter + " ";
                        res.animClick += getItemByName(selects, 'direction_C').getValue();
                    }
                    // get OnHover data
                    animation = getItemByName(selects, 'anim_O').getValue();
                    if (animation !== 'none') {
                        res.animOver = animation + " ";
                        res.animOver += getItemByName(inputs, 'duration_O').getValue() + "s ";
                        res.animOver += getItemByName(selects,'timing_O').getValue() + " ";
                        res.animOver += getItemByName(inputs, 'delay_O').getValue() + "s ";
                        animIter = getItemByName(inputs, 'iteration_O').getValue();
                        if (animIter === "0") {
                            animIter = "infinite";
                        }
                        res.animOver += animIter + " ";
                        res.animOver += getItemByName(selects, 'direction_O').getValue();
                    }
                    //console.log(inputs, selects, res);
                    CKEDITOR.plugins.cssanim.runAddAnimElem(attachedElem, res);
                }
            }]
        }],
//        onLoad: function () {
//            //console.log("ON LOAD cssanimAddAnimDialog", attachedElem);
//        },
        onShow: function () {
            attachedElem = CKEDITOR.plugins.cssanim.curSelectedElement;
            if (CKEDITOR.plugins.cssanim.managePending) {
                //console.log('Manage Pendings is ON !!!!!!!!!!!!!!!!!');
                pendingChangesObj = CKEDITOR.plugins.cssanim.pendingChanges;
                //console.log("ON SHOW PENDING ELEMENTS ============== ", pendingChangesObj);
            }
            this.setupContent();
            var dial = this.getContentElement('mainPanel', 'htmlAnim');
            var inputs = dial.getInputElement();
            var elms = inputs.getElementsByTag('select');
            var elm;
            elm = getItemByName(elms, 'anim_L');
            animSelectionChange(elm.$);
            elm = getItemByName(elms, 'anim_C');
            animSelectionChange(elm.$);
            elm = getItemByName(elms, 'anim_O');
            animSelectionChange(elm.$);

        },
        onOk: function () {
            //console.log("ON OK cssanimAddAnimDialog");
            this.commitContent();
            if (CKEDITOR.plugins.cssanim.managePending) {
                //console.log("ON OK PENDING ELEMENTS ============== ", pendingChangesObj);
                CKEDITOR.plugins.cssanim.pendingChanges = pendingChangesObj;
                CKEDITOR.plugins.cssanim.refreshCssAnimDialogAnimationsTab();
            }
        },
//        onHide: function () {
//            console.log("ON HIDE cssanimAddAnimDialog");
//        },
        onCancel: function () {
            //console.log("ON CANCEL cssanimAddAnimDialog");
            this.foreach(function (widget) {
//                // Make sure IE triggers "change" event on last focused input before closing the dialog. (#7915)
//                if (CKEDITOR.env.ie && (this._.currentFocusIndex == widget.focusIndex)) widget.getInputElement().$.blur();
                if (widget.cancel) {
                	widget.cancel.apply(widget);
                }
            });
            //console.log("CANCEL PENDING ELEMENTS ============== ", pendingChangesObj);
        }
    };
});
