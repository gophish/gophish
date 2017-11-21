/**
 * @author: Önder Ceylan <onderceylan@gmail.com>
 * @copyright Copyright (c) 2013 - Önder Ceylan. All rights reserved.
 * @license Licensed under the terms of GPL, LGPL and MPL licenses.
 * @version 1.0
 *
 * Date: 5/10/13
 * Time: 9:45 AM
 */

// define a prototype toUpperCase fn for Turkish character recognization
String.prototype.trToUpperCase = function(){
    var string = this;
    var letters = { "i": "İ", "ş": "Ş", "ğ": "Ğ", "ü": "Ü", "ö": "Ö", "ç": "Ç", "ı": "I" };
    string = string.replace(/(([iışğüçö]))/g, function(letter){ return letters[letter]; });
    if (typeof(String.prototype.toLocaleUpperCase()) != 'undefined') {
        return string.toLocaleUpperCase();
    } else {
        return string.toUpperCase();
    }
};

// define prototype toLowerCase fn for Turkish character recognization
String.prototype.trToLowerCase = function(){
    var string = this;
    var letters = { "İ": "i", "I": "ı", "Ş": "ş", "Ğ": "ğ", "Ü": "ü", "Ö": "ö", "Ç": "ç" };
    string = string.replace(/(([İIŞĞÜÇÖ]))/g, function(letter){ return letters[letter]; });
    if (typeof(String.prototype.toLocaleLowerCase()) != 'undefined') {
        return string.toLocaleLowerCase();
    } else {
        return string.toLowerCase();
    }
};

// set CKeditor lang
CKEDITOR.plugins.setLang( 'texttransform', 'tr', {
    transformTextSwitchLabel: 'Harf Düzenini Değiştir',
    transformTextToLowercaseLabel: 'Küçük Harfe Dönüştür',
    transformTextToUppercaseLabel: 'Büyük Harfe Dönüştür',
    transformTextCapitalizeLabel: 'Baş Harfleri Büyüt'
});
