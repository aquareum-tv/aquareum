
#include <gst/gst.h>
#ifdef HAVE_CONFIG_H
#include "config.h"
#endif

GST_PLUGIN_STATIC_DECLARE(gofilesink);

void gst_init_gofilesink (void)
{
  GST_PLUGIN_STATIC_REGISTER(gofilesink);
}
