(function (exports) {
    'use strict';

    vis.defineObject('label', {
        createContent: function () {
            this._elem = document.createElement('span');
            this._elem.innerHTML = this.properties.content;
            return this._elem;
        }
    });

})(window);
