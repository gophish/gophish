var page2imagesIframeEditor=null;
CKEDITOR.plugins.add( 'page2images',
{
	icons: 'page2images',
	init : function( editor )
	{
		var pluginName = 'page2images';
		var title = "";
		var validate = "";
		var serverUrl="http://www.page2images.com";
		className="cke_button_page2images";
		var eventMethod = window.addEventListener ? "addEventListener" : "attachEvent";
				var eventer = window[eventMethod];
				var messageEvent = eventMethod == "attachEvent" ? "onmessage" : "message";
				eventer(messageEvent, function (e) {
					  if (e.origin == serverUrl&&e.data.indexOf("'")<0&&e.data.indexOf("tmp/")>-1) {
							editor.insertHtml("<img style='max-width:100%' src=\""+e.data+"\" />");
							// $("#page2image_plugin").remove();
							// $("#page2image_lightboxOverlay").remove();
							var oP = document.getElementById("page2image_plugin");
							oP.parentNode.removeChild(oP);
							var oPL = document.getElementById("page2image_lightboxOverlay");
							oPL.parentNode.removeChild(oPL);
					  }
				}, false);  
		editor.ui.addButton( 'page2images',
			{
				label : title,
				command : pluginName,
				click:function(){
					document.documentElement.scrollTop=0;

					var str = "<div class='page2images_loading' style='display:none;font-weight:bold;position:absolute;left:30px;top:5px;'>Loading...</div><div id='page2images_close' style='cursor:pointer;font-weight:bold;position:absolute;right:20px;top:5px;'><img src='http://www.page2images.com/resources/img/tools/icon_failed.png' /></div>";
					str += "<iframe id='page2image_iframe'   height='100%' width='100%' name='page2image_iframe' border=0 src='"+serverUrl+"/URL-Live-Website-Screenshot-Generator?noDisplayHeaderFooterHTML=1' style='background:#fff;border:0px;' /></div>";
					//if($("#page2image_iframe").length == 0){
					//	$('body').append(str);
					//}
					var oP = document.getElementById("page2image_plugin");
					if(!oP)
					{
						
						
						var newDiv = document.createElement("div");
						newDiv.id="page2image_plugin"; 
						newDiv.style.background = "#fff";
						newDiv.style.position = "absolute";
						newDiv.style.top = "0";
						newDiv.style.width = "1032px";
						var clintW=document.body.clientWidth;
						if(clintW>1032)
						{
							newDiv.style.marginLeft = (clintW-1032)/2+"px";
						}
						newDiv.style.height = "96%";   
						newDiv.style.marginTop = "2%";
						newDiv.style.zIndex = "9000";   
						newDiv.innerHTML = str;
						document.body.insertBefore(newDiv, document.body.childNodes[0]);
						var newOp=document.createElement("div");
						newOp.id="page2image_lightboxOverlay"; 
						newOp.className = "lightboxOverlay"; 
						newOp.style.background = "#000"; 
						newOp.style.opacity = "0.8"; 
						newOp.style.position = "absolute"; 
						newOp.style.zIndex = "1999"; 
						newOp.style.top = "0"; 
						newOp.style.left = "0"; 
						newOp.style.width="100%";
						newOp.style.height=document.body.scrollHeight+"px";
						document.body.insertBefore(newOp, document.body.childNodes[0]);
					}
					var page2images_close=document.getElementById("page2images_close");
					page2images_close.onclick = function(){
						    var oP = document.getElementById("page2image_plugin");
							oP.parentNode.removeChild(oP);
							var oPL = document.getElementById("page2image_lightboxOverlay");
							oPL.parentNode.removeChild(oPL);
					 };
				
					
					
				},
				className:className
			});
	}
} );
/**
 *  Padding text to set off the image in preview area.
 * @name CKEDITOR.config.image_previewText
 * @type String
 * @default "Lorem ipsum dolor..." placehoder text.
 * @example
 * config.image_previewText = CKEDITOR.tools.repeat( '___ ', 100 );
 */
