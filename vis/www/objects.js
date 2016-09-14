(function (exports) {
    'use strict';

    var ObjectBase = Class({
        constructor: function (props, world) {
            this.properties = props;
            this.world = world;
        },

        render: function (elem, outer) {
            this.outerElem = elem;
            this.outer = outer;
        },

        measure: function () {
            return vis.measure(this.properties);
        }
    });

    var CanvasObject = Class(ObjectBase, {
        constructor: function () {
            ObjectBase.prototype.constructor.apply(this, arguments);
        },

        render: function () {
            ObjectBase.prototype.render.apply(this, arguments);
            this._canvas = document.createElement('canvas');
            this._canvas.classList.add('object');
            this._canvas.classList.add(this.properties.type);
            this.outerElem.appendChild(this._canvas);
        },

        place: function (viewRc) {
            this._canvas.setAttribute('width', viewRc.w);
            this._canvas.setAttribute('height', viewRc.h);
            this.paint(viewRc, this._canvas);
        },

        paint: function () {
            // to be overridden
        }
    });

    function defineObject (type, prototype, baseClass) {
        if (baseClass == null) {
            baseClass = ObjectBase;
        }
        var ctor = prototype.constructor;
        prototype.constructor = function () {
            baseClass.prototype.constructor.apply(this, arguments);
            if (ctor != null) {
                ctor.apply(this.arguments);
            }
        };
        var cls = Class(baseClass, prototype, { statics: { Type: type } });
        vis.registerClass(cls);
        return cls;
    }

    function defineCanvasObject (type, prototype) {
        return defineObject(type, prototype, CanvasObject);
    }

    vis.ObjectBase = ObjectBase;
    vis.CanvasObject = CanvasObject;
    vis.defineObject = defineObject;
    vis.defineCanvasObject = defineCanvasObject;
})(window);
