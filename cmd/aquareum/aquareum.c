#include <string.h>
#ifdef __linux__
#include <mistserver.h>
#endif
#include "aquareum.h"

int main(int argc, char *argv[])
{
  if (argc < 2)
  {
    AquareumMain();
  }
#ifdef __linux__
  else if (strncmp("Mist", argv[1], 4) == 0)
  {
    return MistServerMain(argc, argv);
  }
#endif
  else
  {
    AquareumMain();
  }
}
