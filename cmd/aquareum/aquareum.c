#include <string.h>
#include <mistserver.h>
#include "aquareum.h"

int main(int argc, char *argv[])
{
  if (argc < 2)
  {
    AquareumMain();
  }
  else if (strncmp("Mist", argv[1], 4) == 0)
  {
    return MistServerMain(argc, argv);
  }
  else
  {
    AquareumMain();
  }
}
